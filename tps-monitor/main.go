package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jpmorganchase/quorum-profiling/tps-monitor/tpsmon"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

var (
	app   = cli.NewApp()
	flags = []cli.Flag{
		tpsmon.ConsensusFlag,
		tpsmon.DebugFlag,
		tpsmon.TpsPortFlag,
		tpsmon.HttpEndpointFlag,
		tpsmon.ReportFileFlag,
		tpsmon.FromBlockFlag,
		tpsmon.ToBlockFlag,
		tpsmon.AwsMetricsEnabledFlag,
		tpsmon.AwsRegionFlag,
		tpsmon.AwsNwNameFlag,
		tpsmon.AwsInstanceFlag,
		tpsmon.PrometheusPortFlag,

		tpsmon.InfluxdbEnabledFlag,
		tpsmon.InfluxdbEndpointFlag,
		tpsmon.InfluxdbTokenFlag,
		tpsmon.InfluxdbOrgFlag,
		tpsmon.InfluxdbBucketFlag,
		tpsmon.InfluxdbPointNameFlag,
		tpsmon.InfluxdbTagsFlag,
	}
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	app.Action = tps
	app.Before = func(c *cli.Context) error {
		log.Info("starting tps monitor")
		return nil
	}
	app.After = func(c *cli.Context) error {
		log.Info("exiting tps monitor")
		return nil
	}
	app.Flags = flags
	app.Usage = "tpsmonitor connects to a geth client (enabled with JSON-RPC endpoint) and monitors the TPS or calculate TPS for given block range"
}

func tps(ctx *cli.Context) error {
	httpendpoint := ctx.GlobalString(tpsmon.HttpEndpointFlag.Name)
	consensus := ctx.GlobalString(tpsmon.ConsensusFlag.Name)
	awsEnabled := ctx.GlobalBool(tpsmon.AwsMetricsEnabledFlag.Name)
	awsRegion := ctx.GlobalString(tpsmon.AwsRegionFlag.Name)
	awsNwName := ctx.GlobalString(tpsmon.AwsNwNameFlag.Name)
	awsInstance := ctx.GlobalString(tpsmon.AwsInstanceFlag.Name)

	prometheusPort := ctx.GlobalInt(tpsmon.PrometheusPortFlag.Name)

	influxdbEnabled := ctx.GlobalBool(tpsmon.InfluxdbEnabledFlag.Name)
	influxdbEndpoint := ctx.GlobalString(tpsmon.InfluxdbEndpointFlag.Name)
	influxdbToken := ctx.GlobalString(tpsmon.InfluxdbTokenFlag.Name)
	influxdbOrg := ctx.GlobalString(tpsmon.InfluxdbOrgFlag.Name)
	influxdbBucket := ctx.GlobalString(tpsmon.InfluxdbBucketFlag.Name)
	influxdbPoint := ctx.GlobalString(tpsmon.InfluxdbPointNameFlag.Name)
	influxdbTags := ctx.GlobalString(tpsmon.InfluxdbTagsFlag.Name)

	debugMode := ctx.GlobalBool(tpsmon.DebugFlag.Name)
	if httpendpoint == "" {
		return errors.New("httpendpoint is empty")
	}

	if consensus == "" || (consensus != "raft" && consensus != "ibft") {
		return errors.New("invalid consensus. should be raft or ibft")
	}

	if debugMode {
		log.SetLevel(log.DebugLevel)
	}
	var awsService *tpsmon.AwsCloudwatchService
	var promethService *tpsmon.PrometheusMetricsService
	var influxdbService *tpsmon.InfluxdbMetricsService
	var err error
	if awsEnabled {
		awsService = tpsmon.NewCloudwatchService(awsRegion, awsNwName, awsInstance)
	}
	if influxdbEnabled {
		if influxdbEndpoint == "" {
			log.Fatalf("influxdb endpoint is empty")
		}
		if influxdbBucket == "" {
			log.Fatalf("influxdb bucket is empty")
		}
		if influxdbPoint == "" {
			log.Fatalf("influxdb point name is empty")
		}
		if influxdbTags == "" {
			log.Fatalf("influxdb tags is empty")
		}
		if influxdbService, err = tpsmon.NewInfluxdbService(influxdbEndpoint, influxdbToken, influxdbOrg, influxdbBucket, influxdbPoint, influxdbTags); err != nil {
			log.Fatalf("failed to create influxdb service err %v", err)
		}
		log.Info("influxdb service created.")
	}

	if prometheusPort > 0 {
		promethService = tpsmon.NewPrometheusMetricsService(prometheusPort)
		log.Info("prometheus service created.")
	}

	fromBlk := ctx.GlobalUint64(tpsmon.FromBlockFlag.Name)
	toBlk := ctx.GlobalUint64(tpsmon.ToBlockFlag.Name)
	if fromBlk > toBlk {
		log.Fatalf("from block is less than to block no")
	}

	isRaft := ctx.GlobalString(tpsmon.ConsensusFlag.Name) == "raft"

	tm := tpsmon.NewTPSMonitor(awsService, promethService, influxdbService, isRaft, ctx.GlobalString(tpsmon.ReportFileFlag.Name),
		fromBlk, toBlk, httpendpoint)
	startTps(tm)
	tpsPort := ctx.GlobalInt(tpsmon.TpsPortFlag.Name)
	tpsmon.NewTPSServer(tm, tpsPort)
	tm.Wait()
	return nil
}

func startTps(monitor *tpsmon.TPSMonitor) {
	if monitor.IfBlockRangeGiven() {
		go monitor.StartTpsForBlockRange()
		return
	}

	monitor.StartTpsForNewBlocksFromChain()
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Error("Interrupt signal caught, shutting down app")

		go monitor.Stop()

		for i := 5; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Warning(fmt.Sprintf("Shutdown in progress, interrupt %d more times to force shutdown", i-1))
			}
		}
		panic("Forced shutodwn: maximum interrupts given")
	}()
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
