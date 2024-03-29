package store

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/tubuarge/GoHammer/config"
	"github.com/tubuarge/GoHammer/logger"
	"github.com/tubuarge/GoHammer/util"
)

var deployQueue *queue.Queue

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

// getStoreInstance returns a Store Instance deployed on the given client.
func (d *DeployClient) getStoreInstance(conn *ethclient.Client, nodeCipher string) (*Store, error) {
	privateKey, err := crypto.HexToECDSA(nodeCipher)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Error while casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := conn.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}
	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
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
		return nil, err
	}

	d.Logger.TestResult.TotalTxCount += 1
	return instance, nil
}

// callSetItem calls the deployed smart contracts SetItem method.
func (d *DeployClient) callSetItem(storeInst *Store, conn *ethclient.Client, nodeCipher string) error {
	privateKey, err := crypto.HexToECDSA(nodeCipher)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("Error while casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := conn.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}
	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	_, err = storeInst.StoreTransactor.SetItem(auth, [32]byte{1}, [32]byte{2})
	if err != nil {
		return err
	}
	d.Logger.TestResult.TotalTxCount += 1
	return nil
}

func (d *DeployClient) DeployTestProfiles(testProfiles []config.TestProfile) {
	testStartTimestamp := time.Now()

	d.Logger.TestResult = &logger.TestResults{
		TestStartTimestamp: testStartTimestamp,
		TotalTxCount:       0,
	}

	for _, profile := range testProfiles {
		if profile.RoundRobin == true {
			d.TestProfileRR(&profile)
			continue
		}
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
		if testProfile.CallContractMethod {
			d.testNodeCallMethod(&node)
		}
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

func (d *DeployClient) TestProfileRR(testProfile *config.TestProfile) {
	var callMethodRRStructList []*callMethodRRStruct

	if testProfile.CallContractMethod {
		callMethodRRStructList = d.getCallMethodRRStructList(testProfile)
	}

	node := testProfile.Nodes[0]

	for _, deployCount := range node.DeployCounts {
		testStartTimestamp := time.Now()

		d.Logger.WriteTestEntry(
			"Started to test.",
			fmt.Sprintf("%d", deployCount),
			testStartTimestamp,
			logger.SeperatorNone,
		)

		for j := 0; j < deployCount; j++ {
			for i := 0; i <= len(testProfile.Nodes)-1; i++ {
				if testProfile.CallContractMethod {
					d.testNodeRRCallMethod(callMethodRRStructList[i])
					continue
				}
				d.testNodeRR(&testProfile.Nodes[i])
			}
		}
	}
}

func (d *DeployClient) testNodeRR(nodeConfig *config.NodeConfig) {
	conn, err := createConn(nodeConfig.URL)
	if err != nil {
		log.Fatalf("Error while creating ETH Client Connection: %v", err)
	}

	deployContract(conn, nodeConfig.Cipher)

	log.Infof("[%s] deployed on.", nodeConfig.Name)
}

// testNodeRRCallMethod is wrapper function that used when running Round Robin and
// callMethod test profile.
func (d *DeployClient) testNodeRRCallMethod(callMethodRRStruct *callMethodRRStruct) {
	d.callSetItem(callMethodRRStruct.storeInst, callMethodRRStruct.ethClient, callMethodRRStruct.nodeCipher)
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
		d.Logger.TestResult.OverallExecutionTime += elapsedTime
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

func (d *DeployClient) testNodeCallMethod(nodeConfig *config.NodeConfig) {
	conn, err := createConn(nodeConfig.URL)
	if err != nil {
		log.Fatalf("Error while creating ETH Client Connection: %v", err)
	}

	storeInst, err := d.getStoreInstance(conn, nodeConfig.Cipher)
	if err != nil {
		log.Fatalf("Error while creating Store Instance: %v", err)
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
			log.Info("Calling SetItem method")
			d.callSetItem(storeInst, conn, nodeConfig.Cipher)
			//d.deployContract(conn, nodeConfig.Cipher)
		}

		log.Infof("Deployed %d transaction on the given node.", deployCount)
		d.Logger.WriteTestEntry(
			"Ended test.",
			fmt.Sprintf("%s - %d", nodeConfig.Name, deployCount),
			time.Now(),
			logger.SeperatorNone,
		)

		elapsedTime := time.Since(testStartTimestamp)
		d.Logger.TestResult.OverallExecutionTime += elapsedTime
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

// callMethodRRStruct contains information for calling callSetItem
type callMethodRRStruct struct {
	ethClient  *ethclient.Client
	storeInst  *Store
	nodeCipher string
}

// getCallMethodRRStructList returns a list of callMethodRRStruct struct that contains
// required values to call callSetItem function.
// if there is a problem occured while creating ethclient or store instance,
// terminates the program.
func (d *DeployClient) getCallMethodRRStructList(testProfile *config.TestProfile) []*callMethodRRStruct {
	var callMethodRRStructList []*callMethodRRStruct

	// create connections and store instance for every node in the test profile and
	// create a callMethoRRStruct, then add this struct to the slice.
	nodes := testProfile.Nodes
	for _, node := range nodes {
		conn, err := createConn(node.URL)
		if err != nil {
			log.Fatalf("Error while creating ETH Client Connection: %v", err)
			return nil
		}

		storeInst, err := d.getStoreInstance(conn, node.Cipher)
		if err != nil {
			log.Fatalf("Error while creating Store Instance: %v", err)
			return nil
		}

		structInst := &callMethodRRStruct{
			ethClient:  conn,
			storeInst:  storeInst,
			nodeCipher: node.Cipher,
		}

		callMethodRRStructList = append(callMethodRRStructList, structInst)
	}

	return callMethodRRStructList
}

func createConn(nodeUrl string) (*ethclient.Client, error) {
	conn, err := ethclient.Dial(nodeUrl)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
