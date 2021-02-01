package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/tubuarge/GoHammer/config"
	"github.com/tubuarge/GoHammer/rpc"
	"github.com/tubuarge/GoHammer/store"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

var (
	app   = cli.NewApp()
	flags = []cli.Flag{
		DeployNodeUrlFlag,
		DeployNodeCipherFlag,
		DeployCountFlag,
		DeployIntervalFlag,
	}

	rpcClient *rpc.RPCClient
)

type ClientStruct struct {
	client *ethclient.Client
	node   config.NodeConfig
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	rpcClient = rpc.NewRPCClient()

	app.Action = gohammer

	app.Before = func(c *cli.Context) error {
		log.Info("Starting GoHammer.")
		return nil
	}

	app.After = func(c *cli.Context) error {
		log.Info("Exiting GoHammer.")
		return nil
	}
	app.Flags = flags
	app.Usage = "GoHammer deploys a smart contract in the given node and interval."
}

func gohammer(ctx *cli.Context) error {
	nodeUrl := ctx.GlobalString(DeployNodeUrlFlag.Name)
	nodeCipher := ctx.GlobalString(DeployNodeCipherFlag.Name)
	deployCount := ctx.GlobalIntSlice(DeployCountFlag.Name)
	deployInterval := ctx.GlobalString(DeployIntervalFlag.Name)

	// check deploy count is at least have one item.
	if len(deployCount) < 1 {
		return errors.New("at least provide one deploy count.")
	}

	// check if given node url is not empty or has a working node.
	if nodeUrl == "" {
		return errors.New("give a node RPC url.")
	} else {
		isNodeUp, err := rpcClient.IsNodeUp(nodeUrl)
		if err != nil {
			return err
		}
		if !isNodeUp {
			return errors.New("given node is not working.")
		}
	}

	// if deploy interval is not given then use default deploy interval.
	if deployInterval == "" {
		deployInterval = "30s"
	}

	interval, err := time.ParseDuration(deployInterval)
	if err != nil {
		return errors.New("given interval can't convert to duration.")
	}

	deploy(nodeUrl, nodeCipher, deployCount, interval)
	return nil
}

func deploy(nodeUrl, nodeCipher string, deployCount []int, deployInterval time.Duration) {
	conn, err := ethclient.Dial(nodeUrl)
	if err != nil {
		log.Fatal(err)
	}

	for _, elem := range deployCount {
		for i := 0; i < elem; i++ {
			deployContract(conn, nodeCipher)
		}
		log.Infof("Deployed %d transaction on the given node.", elem)
		time.Sleep(deployInterval)
	}
}

func deployContract(conn *ethclient.Client, nodeCipher string) {
	privateKey, err := crypto.HexToECDSA(nodeCipher)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := conn.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, instance, err := store.DeployStore(auth, conn, input)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Address: ", address.Hex())
	log.Info("Tx Hash: ", tx.Hash().Hex())

	_ = instance
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
