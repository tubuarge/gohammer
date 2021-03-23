package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

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

	cfg config.Config
)

type ClientStruct struct {
	client *ethclient.Client
	node   config.NodeConfig
}

func readConfig(cfg *config.Config, configFileName string) {
	configFileName, _ = filepath.Abs(configFileName)
	log.Infof("Loading config: %v", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
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
	app.Usage = "GoHammer deploys a smart contract in the given node(s) and interval according to test profile JSON file given by the user."
}

// checkNodes calls isNodeUp function to ensure that every node is running
// before starting the test.
// If there is a failed node then terminates the GoHammer.
func checkNodes(cfg *config.Config) {
	isOK := true

	profiles := cfg.TestProfiles

	for _, profile := range profiles {
		nodes := profile.Nodes
		for _, node := range nodes {
			isNodeUp, err := rpcClient.IsNodeUp(node.URL)
			if err != nil {
				isOK = false
				log.Errorf("%s node is not running: %v", node.Name, err)
				continue
			}
			if !isNodeUp {
				isOK = false
				log.Errorf("%s node is not running.", node.Name)
				continue
			}
			log.Infof("%s node is OK.", node.Name)
		}
	}

	if !isOK {
		log.Fatalf("Make sure every node given in the test-profile file is running.")
	}
}

func gohammer(ctx *cli.Context) error {
	testProfileFileName := ctx.GlobalString(TestProfileConfigFileFlag.Name)

	// check if test profile name is not empty
	if testProfileFileName != "" {
		return errors.New("Please, enter a test-profile file")
	}

	readConfig(&cfg, testProfileFileName)
	checkNodes(&cfg)

	startTest(&cfg)
	return nil
}

func startTest(cfg *config.Config) {
	profiles := cfg.TestProfiles

	if cfg.Concurrent {
		for _, profile := range profiles {
			nodes := profile.Nodes

			for _, node := range nodes {

			}
		}
	}

}

func deployTestProfile(testProfile *config.TestProfile) error {
	for _, node := range testProfile.Nodes {
		for _, elemDeployCount := range node.DeployCount {
			deploy(node.URL, node.Cipher, elemDeployCount)
		}
	}
}

func deploy(nodeUrl, nodeCipher string, deployCount int) {
	conn, err := ethclient.Dial(nodeUrl)
	if err != nil {
		log.Fatal(err)
	}

	deployContract(conn, nodeCipher)
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
