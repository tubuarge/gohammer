package tpsmon

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/jpmorganchase/quorum-profiling/tps-monitor/reader"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// create new dummy block data
func getNewBlockData(n int64, txnCnt int, isRaft bool, t1 time.Time) *reader.BlockData {
	var tm int64
	if isRaft {
		tm = t1.UnixNano()
	} else {
		tm = t1.Unix()
	}
	blkData := &reader.BlockData{
		Number:   uint64(n),
		GasLimit: uint64(0),
		GasUsed:  uint64(0),
		Time:     uint64(tm),
		TxnCnt:   txnCnt,
	}
	return blkData
}

func testTPS(isRaft bool, t *testing.T) {
	assert := assert.New(t)
	var tempFile string
	if f, err := ioutil.TempFile("", "_tps_test"); err != nil {
		t.Fatalf("error creating temp file for test")
	} else {
		tempFile = f.Name()
	}

	defer os.Remove(tempFile)

	tm := NewTPSMonitor(nil, nil, nil, isRaft, tempFile, 1, 10, "")

	assert.NotNil(tm, "creating tps monitor failed")

	tm.init()

	var blksArr []*reader.BlockData
	var c int
	var t1, t2 time.Time
	if isRaft {
		unsec := time.Now().UnixNano()
		nsec := unsec % 1e9
		t1 = time.Unix(unsec/1e9, nsec)
	} else {
		t1 = time.Unix(time.Now().Unix(), 0)
	}
	for c = 1; c <= 20; c++ {
		t2 = t1.Add(time.Second) //.Add(time.Second)
		blksArr = append(blksArr, getNewBlockData(int64(c), c*1000, isRaft, t1))
		t1 = t2
	}

	for _, b := range blksArr {
		tm.readBlock(b)
	}
	tm.Stop()

	tm.printTPS()
	var expTxnCnt uint64 = 190000
	var expBlkCnt uint64 = 19
	var expTps uint32 = 10000
	expTpsRecs := 19

	lr := len(tm.tpsRecs)

	assert.Equal(len(tm.tpsRecs), expTpsRecs, "tps record count mismatch")

	txnCnt := tm.tpsRecs[lr-1].txns
	blkCnt := tm.tpsRecs[lr-1].blks
	tps := tm.tpsRecs[lr-1].tps

	assert.Equal(txnCnt, expTxnCnt, "total txn count mismatch")
	assert.Equal(blkCnt, expBlkCnt, "block count mismatch")
	assert.Equal(tps, expTps, "tps mismatch")

	if tpsFile, err := os.Open(tempFile); err != nil {
		t.Errorf("opening tps file %s failed", tempFile)
	} else {
		lineCnt := 0
		expLineCnt := 20
		firstLine := "localTime,refTime,TPS,TxnCount,BlockCount"
		lineStr := ""
		var lineStrArr []string
		scanner := bufio.NewScanner(tpsFile)
		for scanner.Scan() {
			lineCnt++
			lineStr = scanner.Text()
			lineStrArr = strings.Split(lineStr, ",")
			if lineCnt == 1 {
				assert.Equal(len(lineStrArr), 5, "tps report file header fields mismatch")
				assert.Equal(lineStr, firstLine, "tps report file header data mismatch")
			}
		}

		assert.Equal(lineCnt, expLineCnt, "tps report file lines")

		if ftps, err := strconv.ParseInt(lineStrArr[2], 10, 32); err != nil {
			t.Errorf("tps data in file is wrong - tps is not a valid number")
		} else {
			assert.Equal(uint32(ftps), expTps, "tps report file - tps")
		}

		if ftxn, err := strconv.ParseInt(lineStrArr[3], 10, 64); err != nil {
			t.Errorf("tps data in file is wrong - tps is not a valid number")
		} else {
			assert.Equal(uint64(ftxn), expTxnCnt, "tps report file - tps")
		}

		if fblk, err := strconv.ParseInt(lineStrArr[4], 10, 64); err != nil {
			t.Errorf("tps data in file is wrong - tps is not a valid number")
		} else {
			assert.Equal(uint64(fblk), expBlkCnt, "tps report file - tps")
		}

	}

}
func TestTPSForIbft(t *testing.T) {
	f, _ := ioutil.TempFile("", "_tps_test")
	fn := f.Name()
	t.Log(fn)
	testTPS(false, t)
}

func TestTPSForRaft(t *testing.T) {
	testTPS(true, t)
}

func TestExecAws(t *testing.T) {
	mySession := session.Must(session.NewSession())
	// Create a CloudWatch chainReader with additional configuration
	svc := cloudwatch.New(mySession, aws.NewConfig().WithRegion("ap-southeast-1"))
	t.Log(svc)
	var pmd *cloudwatch.PutMetricDataInput
	var mdn *cloudwatch.MetricDatum
	dname := "Instance"
	dvalue := "NodeX"
	nspace := "stX-q24-X.Y.Z"
	mname := "TPS"
	var tps uint64 = 1000
	var value float64 = float64(tps)
	ts := time.Now()
	dimension := &cloudwatch.Dimension{Name: &dname, Value: &dvalue}
	mdn = &cloudwatch.MetricDatum{
		Dimensions: []*cloudwatch.Dimension{dimension},
		MetricName: &mname,
		Timestamp:  &ts,
		Value:      &value,
	}
	pmd = &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{mdn},
		Namespace:  &nspace,
	}
	svc.PutMetricData(pmd)
}
