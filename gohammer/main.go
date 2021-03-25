package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"

	"github.com/tubuarge/GoHammer/config"
	"github.com/tubuarge/GoHammer/logger"
	"github.com/tubuarge/GoHammer/rpc"
	"github.com/tubuarge/GoHammer/store"
)

const (
	TestResultFilename = "result.log"
)

var (
	app   = cli.NewApp()
	flags = []cli.Flag{
		TestProfileConfigFileFlag,
	}

	rpcClient    *rpc.RPCClient
	loggerClient *logger.LogClient
	deployClient *store.DeployClient

	cfg config.Config
)

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

	loggerClient, err := logger.NewLogClient(TestResultFilename)
	if err != nil {
		log.Error("Couln't create log file: %v", err)
	}

	deployClient = store.NewDeployClient(loggerClient)

	rpcClient = rpc.NewRPCClient()

	app.Action = gohammer

	app.Before = func(c *cli.Context) error {
		log.Info("Starting GoHammer.")
		return nil
	}

	app.After = func(c *cli.Context) error {
		loggerClient.CloseFile()

		log.Info("Exiting GoHammer.")
		return nil
	}
	app.Flags = flags
	app.Usage = "GoHammer deploys a smart contract in the given node(s) and interval according to test profile JSON file that is given by the user."
}

func gohammer(ctx *cli.Context) error {
	testProfileFileName := ctx.GlobalString(TestProfileConfigFileFlag.Name)

	// check if test profile name is not empty
	if testProfileFileName == "" {
		return errors.New("Please, enter a test-profile file: --testprofilefile <file.json>")
	}

	readConfig(&cfg, testProfileFileName)
	rpcClient.CheckNodes(&cfg)

	testProfiles := cfg.TestProfiles
	startTest(testProfiles)

	return nil
}

func startTest(testProfiles []config.TestProfile) {
	for _, profile := range testProfiles {
		deployClient.DeployTestProfile(&profile)
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
