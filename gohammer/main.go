package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"./config"
	"./rpc"

	log "github.com/sirupsen/logrus"
)

const (
	node1RPCAddress = "http://localhost:22000"
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

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
