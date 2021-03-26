package store

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/tubuarge/GoHammer/config"
	"github.com/tubuarge/GoHammer/logger"
	"github.com/tubuarge/GoHammer/util"
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
	//address, tx, instance, err := DeployStore(auth, conn, input)
	_, _, instance, err := DeployStore(auth, conn, input)
	if err != nil {
		log.Fatal(err)
	}

	/*
		log.Info("Address: ", address.Hex())
		log.Info("Tx Hash: ", tx.Hash().Hex())
	*/

	_ = instance
}

func (d *DeployClient) DeployTestProfiles(testProfiles []config.TestProfile) {
	testStartTimestamp := time.Now()

	d.Logger.TestResult = &logger.TestResults{
		TestStartTimestamp: testStartTimestamp,
		TotalTxCount:       0,
	}

	for _, profile := range testProfiles {
		d.TestProfile(&profile)
	}

	testEndTimestamp := time.Now()
	elapsedTime := time.Since(testStartTimestamp)

	d.Logger.TestResult.TestStartTimestamp = testStartTimestamp
	d.Logger.TestResult.TestEndTimestamp = testEndTimestamp
	d.Logger.TestResult.OverallExecutionTime = elapsedTime
}

func (d *DeployClient) TestProfile(testProfile *config.TestProfile) {
	log.Infof("Starting to test [%s]...", testProfile.Name)

	testStartTimestamp := time.Now()
	d.Logger.WriteTestEntry(
		"Started to test.",
		testProfile.Name,
		testStartTimestamp,
		logger.SeperatorNewLine,
	)

	for _, node := range testProfile.Nodes {
		log.Infof("Starting to deploy on [%s] node...", node.Name)
		d.testNode(&node)
	}

	d.Logger.WriteTestEntry(
		"Ended test.",
		testProfile.Name,
		time.Now(),
		logger.SeperatorNone,
	)

	elapsedTime := time.Since(testStartTimestamp)
	d.Logger.WriteTestEntry(
		fmt.Sprintf("Elapsed test run time: %s", elapsedTime),
		testProfile.Name,
		time.Now(),
		logger.SeperatorProfile,
	)
}

func (d *DeployClient) testNode(nodeConfig *config.NodeConfig) {
	conn, err := createConn(nodeConfig.URL)
	if err != nil {
		log.Fatalf("Error while creating ETH Client Connection: %v", err)
	}

	for _, deployCount := range nodeConfig.DeployCounts {
		testStartTimestamp := time.Now()
		d.Logger.WriteTestEntry(
			"Started to test.",
			fmt.Sprintf("%s - %d", nodeConfig.Name, deployCount),
			testStartTimestamp,
			logger.SeperatorNone,
		)

		for i := 0; i < deployCount; i++ {
			deployContract(conn, nodeConfig.Cipher)
			d.Logger.TestResult.TotalTxCount++
		}

		log.Infof("Deployed %d transaction on the given node.", deployCount)
		d.Logger.WriteTestEntry(
			"Ended test.",
			fmt.Sprintf("%s - %d", nodeConfig.Name, deployCount),
			time.Now(),
			logger.SeperatorNone,
		)

		elapsedTime := time.Since(testStartTimestamp)
		d.Logger.WriteTestEntry(
			fmt.Sprintf("Elapsed test run time: %s", elapsedTime),
			fmt.Sprintf("%s - %d", nodeConfig.Name, deployCount),
			time.Now(),
			logger.SeperatorNewLine,
		)

		duration, err := util.ParseDuration(nodeConfig.DeployInterval)
		if err != nil {
			log.Errorf("Error while parsing deploy intervar: %v", err)
		}
		time.Sleep(duration)
	}
}

func createConn(nodeUrl string) (*ethclient.Client, error) {
	conn, err := ethclient.Dial(nodeUrl)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
