package tpsmon

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jpmorganchase/quorum-profiling/tps-monitor/reader"
)

// TPSRecord represents a data point of TPS at a specific time
type TPSRecord struct {
	rtime string // reference time starts from time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	ltime string // block time in local time
	tps   uint32 // no of transactions per second
	blks  uint64 // total block count
	txns  uint64 // total transaction count
}

func (t TPSRecord) String() string {
	return fmt.Sprintf("TPSRecord: ltime:%v rtime:%v tps:%v txns:%v blks:%v", t.ltime, t.rtime, t.tps, t.txns, t.blks)
}

func (t TPSRecord) ReportString() string {
	return fmt.Sprintf("%v,%v,%v,%v,%v\n", t.ltime, t.rtime, t.tps, t.txns, t.blks)
}

// TPSMonitor implements a monitor service
type TPSMonitor struct {
	isRaft          bool                   // represents consensus
	bdCh            chan *reader.BlockData // block data from chain
	chainReader     *reader.GethClient     // ethereum chainReader
	tpsRecs         []TPSRecord            // list of TPS data points recorded
	report          string                 // report name to store TPS data points
	fromBlk         uint64                 // from block number
	toBlk           uint64                 // to block number
	stopc           chan struct{}          // stop channel
	firstBlkTime    *time.Time             // first block's time
	refTime         time.Time              // reference time
	refTimeNext     time.Time              // next expected reference time
	blkTimeNext     time.Time              // next expected block time
	blkCnt          uint64
	txnsCnt         uint64   // total transaction count
	rptFile         *os.File // report file
	awsService      *AwsCloudwatchService
	promethService  *PrometheusMetricsService
	influxdbService *InfluxdbMetricsService
}

// Date format to show only hour and minute
const (
	dateFmtMinSec = "02 Jan 2006 15:04:05"
)

func NewTPSMonitor(awsService *AwsCloudwatchService, promethService *PrometheusMetricsService, influxdbService *InfluxdbMetricsService, isRaft bool, report string, frmBlk uint64, toBlk uint64, httpendpoint string) *TPSMonitor {
	bdCh := make(chan *reader.BlockData, 1)
	tm := &TPSMonitor{
		isRaft:          isRaft,
		report:          report,
		bdCh:            bdCh,
		fromBlk:         frmBlk,
		toBlk:           toBlk,
		stopc:           make(chan struct{}),
		awsService:      awsService,
		promethService:  promethService,
		influxdbService: influxdbService,
	}
	tm.chainReader = reader.NewGethClient(httpendpoint, bdCh, tm.stopc)
	if tm.report != "" {
		var err error
		if tm.rptFile, err = os.Create(tm.report); err != nil {
			log.Fatalf("error creating report file %s\n", tm.report)
		}
		if _, err := tm.rptFile.WriteString("localTime,refTime,TPS,TxnCount,BlockCount\n"); err != nil {
			log.Errorf("writing to report failed err:%v", err)
		}
		tm.rptFile.Sync()
	}
	return tm
}

func (tm *TPSMonitor) IfBlockRangeGiven() bool {
	return tm.fromBlk > 0 && tm.toBlk > 0
}

func (tm *TPSMonitor) StartTpsForBlockRange() {
	tm.init()
	log.Infof("tps calc started - fromBlock:%v toBlock:%v\n", tm.fromBlk, tm.toBlk)
	tm.calcTpsFromBlockRange()
}

// starts service to calculate tps
func (tm *TPSMonitor) StartTpsForNewBlocksFromChain() {
	tm.init()
	if tm.promethService != nil {
		go tm.promethService.Start()
		log.Infof("prometheus service started")
	}
	go tm.chainReader.Start()
	go tm.calcTpsFromNewBlocks()
	log.Infof("tps monitor started")
}

// stops service
func (tm *TPSMonitor) Stop() {
	close(tm.stopc)
	if tm.rptFile != nil {
		tm.rptFile.Close()
	}
}

// waits for stop signal to end service
func (tm *TPSMonitor) Wait() {
	log.Infof("tps monitor waiting to stop")
	<-tm.stopc
	log.Infof("tps monitor wait over stopping")
}

// initializes service
func (tm *TPSMonitor) init() {
	if tm.isRaft {
		log.Infof("consensus is raft")
	} else {
		log.Infof("consensus is ibft")
	}
	tm.refTime = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	tm.refTimeNext = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	tm.blkCnt = 0
	tm.txnsCnt = 0
}

// read data from block and calculated TPS
func (tm *TPSMonitor) readBlock(block *reader.BlockData) {
	var blkTime time.Time
	var tps uint64
	var nilBlk = false
	// triggered by ticker to generate periodic TPS
	if block == nil {
		blkTime = time.Now()
		nilBlk = true
	} else {
		if tm.isRaft {
			r := block.Time % 1e9
			blkTime = time.Unix(int64(block.Time/1e9), int64(r))
		} else {
			blkTime = time.Unix(int64(block.Time), 0)
		}
	}

	if tm.firstBlkTime != nil {
		totSecs := blkTime.Sub(*tm.firstBlkTime).Milliseconds() / 1000
		log.Debugf("total secs:%v total tx:%v", totSecs, tm.txnsCnt)
		if totSecs > 0 {
			tps = tm.txnsCnt / uint64(totSecs)
			log.Infof("TPS:%v txnsCnt:%v blkCnt:%v", tps, tm.txnsCnt, tm.blkCnt)
		}
	}

	if tm.firstBlkTime == nil {
		tm.firstBlkTime = &blkTime
		tm.refTimeNext = tm.refTimeNext.Add(time.Second)
		tm.blkTimeNext = blkTime.Add(time.Second)
	}

	// report tps to file and aws clould watch every second
	if blkTime.After(tm.blkTimeNext) || blkTime.Equal(tm.blkTimeNext) {
		ltime := tm.blkTimeNext.Format(dateFmtMinSec)
		yd := tm.refTimeNext.YearDay() - 1
		hh := tm.refTimeNext.Hour()
		mm := tm.refTimeNext.Minute()
		ss := tm.refTimeNext.Second()
		rtime := fmt.Sprintf("%02d:%02d:%02d:%02d", yd, hh, mm, ss)

		tr := TPSRecord{rtime: rtime, ltime: ltime, tps: uint32(tps), blks: tm.blkCnt, txns: tm.txnsCnt}
		log.Debug(tr.String())
		if tm.rptFile != nil {
			if _, err := tm.rptFile.WriteString(tr.ReportString()); err != nil {
				log.Errorf("writing to report failed %v", err)
			}
			tm.rptFile.Sync()
		}
		tm.tpsRecs = append(tm.tpsRecs, tr)
		//publish metrics to aws cloudwatch
		go tm.putMetricsInAws(tm.blkTimeNext, fmt.Sprintf("%v", tps), fmt.Sprintf("%v", tm.txnsCnt), fmt.Sprintf("%v", tm.blkCnt))
		//publish metrics to prometheus
		go tm.putMetricsInPrometheus(tm.blkTimeNext, tps, tm.txnsCnt, tm.blkCnt)
		//publish metrics to influxdb
		go tm.putMetricsInInfluxdb(tm.blkTimeNext, tps, tm.txnsCnt, tm.blkCnt)
		tm.refTimeNext = tm.refTimeNext.Add(time.Second)
		tm.blkTimeNext = tm.blkTimeNext.Add(time.Second)
	}

	if !nilBlk {
		tm.blkCnt++
		tm.txnsCnt += uint64(block.TxnCnt)
	}
}

func (tm *TPSMonitor) putMetricsInAws(lt time.Time, tps string, txnCnt string, blkCnt string) {
	if tm.awsService != nil {
		tm.awsService.PutMetrics("TPS", tps, lt)
		tm.awsService.PutMetrics("TxnCount", txnCnt, lt)
		tm.awsService.PutMetrics("BlockCount", blkCnt, lt)
	}
}

// calculates TPS for new block added to the chain
func (tm *TPSMonitor) calcTpsFromNewBlocks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case block := <-tm.bdCh:
			if block == nil {
				log.Fatal("reading block data failed")
				return
			}
			log.Infof("received new block %v", block)
			tm.readBlock(block)
		case <-ticker.C:
			tm.readBlock(nil)
		case <-tm.stopc:
			log.Warning("tps monitor stopped - exit loop")
			return
		}
	}
}

// calculates TPS for a given block range
func (tm *TPSMonitor) calcTpsFromBlockRange() {
	stBlk := tm.fromBlk
	toBlk := tm.toBlk
	for stBlk <= toBlk {
		block, err := tm.chainReader.GetBlock(stBlk)
		if err != nil {
			log.Fatal(err)
		}
		tm.readBlock(block)
		stBlk++
	}
}

func (tm *TPSMonitor) printTPS() {
	trl := len(tm.tpsRecs)
	log.Infof("Total tps records %d", trl)
	for i, v := range tm.tpsRecs {
		log.Infof("%d. %v", i, v.String())
	}
}

func (tm *TPSMonitor) putMetricsInPrometheus(tmRef time.Time, tps uint64, txnCnt uint64, blkCnt uint64) {
	if tm.promethService != nil {
		tm.promethService.publishMetrics(tmRef, tps, txnCnt, blkCnt)
	}
}

func (tm *TPSMonitor) putMetricsInInfluxdb(t time.Time, tps uint64, txnsCnt uint64, blkCnt uint64) {
	if tm.influxdbService != nil {
		tm.influxdbService.PushMetrics(t, tps, txnsCnt, blkCnt)
	}
}
