package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"./config"
	"./rpc"
	"./token"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

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

func main() {
	readConfig(&cfg, "config.json")
	rpcClient := rpc.NewRPCClient()

	for _, node := range cfg.Nodes {
		isNodeUp, err := rpcClient.IsNodeUp(node.URL)
		if err != nil {
			log.Error(err)
		}

		if isNodeUp {
			log.Infof("'%s' is 'UP'.", node.Name)
		} else {
			log.Infof("'%s' is 'NOT UP'", node.Name)
		}
	}

	conn, err := ethclient.Dial(cfg.Nodes[0].URL)
	if err != nil {
		log.Fatal(err)
	}

	newToken, err := token.NewToken(common.HexToAddress("0xed9d02e382b34818e88b88a309c7fe71e65f419d"), conn)
	if err != nil {
		log.Fatal(err)
	}

	name, err := newToken.Name(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Token name:", name)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
