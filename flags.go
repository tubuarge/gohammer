package main

import "gopkg.in/urfave/cli.v1"

var (
	DeployNodeUrlFlag = cli.StringFlag{
		Name:  "nodeurl",
		Usage: `Node RPC URL`,
	}

	DeployNodeCipherFlag = cli.StringFlag{
		Name:  "nodecipher",
		Usage: `Node cipher`,
	}

	DeployCountFlag = cli.IntSliceFlag{
		Name:  "deploycount",
		Usage: `Deploy count`,
	}

	DeployIntervalFlag = cli.StringFlag{
		Name:  "deployinterval",
		Usage: `Deploy interval`,
	}
)
