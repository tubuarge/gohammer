package main

import "gopkg.in/urfave/cli.v1"

var (
	DeployCountFlag = cli.IntFlag{
		Name:  "deploycount",
		Usage: `Number of deploys of the contract`,
	}

	DeployNodeUrlFlag = cli.StringFlag{
		Name:  "nodeurl",
		Usage: `node url where the contract will be deployed`,
	}

	DeployNodeCipherFlag = cli.StringFlag{
		Name:  "nodecipher",
		Usage: `node cipher to call contract on the node`,
	}
)
