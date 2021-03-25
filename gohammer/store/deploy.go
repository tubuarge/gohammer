package store

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/tubuarge/GoHammer/config"
	"github.com/tubuarge/GoHammer/logger"
)

type DeployClient struct {
	Logger *logger.LogClient
}

func NewDeployClient(logClient *logger.LogClient) *DeployClient {
	return &DeployClient{Logger: logClient}
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
	address, tx, instance, err := DeployStore(auth, conn, input)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Address: ", address.Hex())
	log.Info("Tx Hash: ", tx.Hash().Hex())

	_ = instance
}

func (d *DeployClient) DeployTestProfile(testProfile *config.TestProfile) {
	log.Infof("Starting to test %s...", testProfile.Name)

	testStartTimestamp := time.Now()
	totalTxCount := 0

	for _, node := range testProfile.Nodes {
		log.Infof("Starting to deploy on %s node", node.Name)
		for _, elemDeployCount := range node.DeployCount {
			deploy(node.URL, node.Cipher, elemDeployCount)
			totalTxCount++
		}
	}
	testEndTimestamp := time.Now()
	elapsedTime := time.Since(testStartTimestamp)

	d.Logger.TestResult.TestStartTimestamp = testStartTimestamp

	d.Logger.TestResult = &logger.TestResults{
		TestStartTimestamp:   testStartTimestamp,
		TestEndTimestamp:     testEndTimestamp,
		OverallExecutionTime: elapsedTime,
		TotalTxCount:         totalTxCount,
	}
}

func deploy(nodeUrl, nodeCipher string, deployCount int) {
	conn, err := createConn(nodeUrl)
	if err != nil {
		log.Fatalf("Error while creating ETH Client Connection: %v", err)
	}

	for i := 0; i < deployCount; i++ {
		deployContract(conn, nodeCipher)
	}
	log.Infof("Deployed %d transaction on the given node.", deployCount)
}

func createConn(nodeUrl string) (*ethclient.Client, error) {
	conn, err := ethclient.Dial(nodeUrl)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
