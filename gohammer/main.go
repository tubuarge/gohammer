package main

import (
	"os"
	"os/signal"
	"syscall"

	"./rpc"

	log "github.com/sirupsen/logrus"
)

const (
	node1RPCAddress = "http://localhost:22000"
)

func isNodeUp() {

}

func sendRPCCall() {

}

func main() {
	rpcClient := rpc.NewRPCClient()

	isNodeUp, err := rpcClient.IsNodeUp(node1RPCAddress)
	if err != nil {
		log.Error(err)
	}

	if isNodeUp {
		log.Infof("'%s' node is 'UP'.", node1RPCAddress)
	} else {
		log.Infof("'%s' node is 'NOT UP'", node1RPCAddress)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
