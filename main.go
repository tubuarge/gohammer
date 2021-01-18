package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tubuarge/GoHammer/config"
	"github.com/tubuarge/GoHammer/rpc"

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
		DeployCountFlag,
		DeployNodeCipherFlag,
	}
)

type ClientStruct struct {
	client *ethclient.Client
	node   config.NodeConfig
}

func init() {
	app.Flags = flags
}

func gohammer() {

}

var cfg config.Config

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

func deployContract(client *ClientStruct) {
	privateKey, err := crypto.HexToECDSA(client.node.Cipher)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := client.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, instance, err := DeployMain(auth, client.client, input)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Address: ", address.Hex())
	log.Info("Tx Hash: ", tx.Hash().Hex())

	_ = instance
}

func deployOnOneNode(client *ClientStruct, deployCount int) {
	start := time.Now()

	for count := 0; count < deployCount; count++ {
		deployContract(client)
	}

	elapsed := time.Since(start)
	log.Infof("Deploying %d contracts took: %s", deployCount, elapsed)
	return
}

func deployOnAllNodes(clients []*ClientStruct, deployCount int) {
	start := time.Now()

	for count := 0; count < deployCount; count++ {
		for _, client := range clients {
			deployContract(client)
		}
	}
	elapsed := time.Since(start)
	log.Infof("Deploying 100 contracts took: %s", elapsed)
	return

}

func deploy() {}

func main() {
	readConfig(&cfg, "config.json")
	rpcClient := rpc.NewRPCClient()

	var clients []*ClientStruct

	for _, node := range cfg.Nodes {
		isNodeUp, err := rpcClient.IsNodeUp(node.URL)
		if err != nil {
			log.Error(err)
		}

		if isNodeUp {
			log.Infof("'%s' is 'UP'.", node.Name)
			conn, err := ethclient.Dial(node.URL)
			if err != nil {
				log.Error(err)
				continue
			}
			clientStruct := &ClientStruct{
				client: conn,
				node:   node,
			}
			clients = append(clients, clientStruct)
			log.Info(clients)
		} else {
			log.Infof("'%s' is 'NOT UP'", node.Name)
		}
	}

	//deployOnAllNodes(clients)
	deployOnOneNode(clients[0], 100)
	time.Sleep(1 * time.Minute)

	deployOnOneNode(clients[0], 500)
	time.Sleep(1 * time.Minute)

	deployOnOneNode(clients[0], 1000)
	time.Sleep(1 * time.Minute)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
