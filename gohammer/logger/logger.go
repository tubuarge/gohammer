package logger

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tubuarge/GoHammer/util"
)

const (
	SeperatorNone    = 0
	SeperatorProfile = 1
	SeperatorNewLine = 2
)

type TestResults struct {
	TestStartTimestamp   time.Time
	TestEndTimestamp     time.Time
	OverallExecutionTime time.Duration

	TotalTxCount int
}

type LogClient struct {
	LogFile    *os.File
	TestResult *TestResults
}

func NewLogClient(logDirFile *os.File) *LogClient {
	return &LogClient{
		LogFile:    logDirFile,
		TestResult: &TestResults{},
	}
}

func CreateLogFile(filepath, filename string) (*os.File, error) {
	var fullPath string
	//if filepath is empty then, user didn't pass logDir cmd option
	//so use default dir.
	if filepath != "" {
		//check if dir exists in the given filepath or not
		//if it is not exits then create the dir
		if !util.IsDirExists(filepath) {
			util.CreateDir(filepath)
		}
		fullPath = filepath + "/"
	}

	logFilename := getLogFilename(filename)
	fullPath += logFilename

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getLogFilename(filename string) string {
	ts := util.GetFormattedTimestampNow()
	return ts + "_" + filename
}

func (l *LogClient) WriteFile(data []byte) error {
	_, err := l.LogFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (l *LogClient) WriteTestEntry(msg, entryTitle string, timestamp time.Time, seperatorType int) {
	seperatorStr := ""

	switch seperatorType {
	case 1:
		seperatorStr = util.GetTestEntrySeperatorStr()
	case 2:
		seperatorStr = "\n"
	default:
		seperatorStr = ""
	}

	entry := fmt.Sprintf("[%s] %s: %s\n%s",
		entryTitle,
		util.GetFormattedTimestamp(util.LoggerTimestampLayout, timestamp),
		msg,
		seperatorStr,
	)

	err := l.WriteFile([]byte(entry))
	if err != nil {
		log.Errorf("Error while writing Test Entry: %v", err)
	}
}

func (l *LogClient) WriteTestResults() error {
	strData := fmt.Sprintf("Test Started At: %v\n"+
		"\t\tTest Ended At: %v\n"+
		"\t\tTotal Test Execution Time: %v\n"+
		"\t\tTotal Transaction Count: %d\n",
		util.GetFormattedTimestamp(util.LoggerTimestampLayout,
			l.TestResult.TestStartTimestamp),
		util.GetFormattedTimestamp(util.LoggerTimestampLayout,
			l.TestResult.TestEndTimestamp),
		fmt.Sprintf("%s", l.TestResult.OverallExecutionTime),
		l.TestResult.TotalTxCount)

	err := l.WriteFile([]byte(strData))
	if err != nil {
		return err
	}
	return nil
}

func (l *LogClient) CloseFile() error {
	err := l.LogFile.Close()
	if err != nil {
		return err
	}
	return nil
}
