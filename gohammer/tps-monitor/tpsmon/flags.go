package tpsmon

import "gopkg.in/urfave/cli.v1"

var (
	AwsMetricsEnabledFlag = cli.BoolFlag{
		Name:  "awsmetrics",
		Usage: `AWS metrics enabled`,
	}

	AwsRegionFlag = cli.StringFlag{
		Name:  "awsregion",
		Usage: `AWS region`,
	}

	AwsNwNameFlag = cli.StringFlag{
		Name:  "awsnetwork",
		Usage: `AWS network name`,
	}

	AwsInstanceFlag = cli.StringFlag{
		Name:  "awsinst",
		Usage: `AWS instance name`,
	}

	InfluxdbEnabledFlag = cli.BoolFlag{
		Name:  "influxdb",
		Usage: `Influxdb enabled`,
	}

	InfluxdbEndpointFlag = cli.StringFlag{
		Name:  "influxdb.endpoint",
		Usage: `Influxdb endpoint`,
		Value: "http://localhost:8086",
	}

	InfluxdbTokenFlag = cli.StringFlag{
		Name:  "influxdb.token",
		Usage: `Influxdb token or username:password`,
		Value: ":",
	}
	InfluxdbOrgFlag = cli.StringFlag{
		Name:  "influxdb.org",
		Usage: `Influxdb org name`,
		Value: "",
	}
	InfluxdbBucketFlag = cli.StringFlag{
		Name:  "influxdb.bucket",
		Usage: `Influxdb bucket or database name`,
		Value: "telegraf",
	}
	InfluxdbPointNameFlag = cli.StringFlag{
		Name:  "influxdb.measurement",
		Usage: `Influxdb measurement name`,
		Value: "quorum_tps",
	}
	InfluxdbTagsFlag = cli.StringFlag{
		Name:  "influxdb.tags",
		Usage: `Influxdb tags (comma separated list of key=value pairs)`,
		Value: "system=quorum,comp=tps",
	}
	ConsensusFlag = cli.StringFlag{
		Name:  "consensus",
		Usage: `Name of consensus ("raft", "ibft")`,
	}

	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: `Debug mode`,
	}

	HttpEndpointFlag = cli.StringFlag{
		Name:  "httpendpoint",
		Usage: "Geth's http endpoint",
	}

	ReportFileFlag = cli.StringFlag{
		Name:  "report",
		Usage: "Full path of csv report file",
		Value: "tps-report.csv",
	}

	TpsPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Http port for tps monitor",
		Value: 7575,
	}

	FromBlockFlag = cli.Uint64Flag{
		Name:  "from",
		Usage: "From block no",
		Value: 0,
	}

	ToBlockFlag = cli.Uint64Flag{
		Name:  "to",
		Usage: "To block no",
		Value: 0,
	}

	PrometheusPortFlag = cli.IntFlag{
		Name:  "prometheusport",
		Usage: "Enable prometheus metrics",
		Value: 0,
	}
)
