package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tubuarge/GoHammer/util"
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

func NewLogClient(filename string) (*LogClient, error) {
	file, err := CreateLogFile(filename)
	if err != nil {
		return nil, err
	}
	return &LogClient{
		LogFile:    file,
		TestResult: &TestResults{},
	}, nil
}

func CreateLogFile(filename string) (*os.File, error) {
	ts := util.GetFormattedTimestampNow()
	fullFilename := ts + " " + filename
	file, err := os.Create(fullFilename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (l *LogClient) WriteFile(data []byte) error {
	_, err := l.LogFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (l *LogClient) WriteTestEntry(msg, entryTitle string, timestamp time.Time) {
	entry := fmt.Sprintf("[%s] %s: %s\n",
		entryTitle,
		util.GetFormattedTimestamp(util.LoggerTimestampLayout, timestamp),
		msg)

	err := l.WriteFile([]byte(entry))
	if err != nil {
		log.Errorf("Error while writing Test Entry: %v", err)
	}
}
func (l *LogClient) WriteNewLine() {
	l.WriteFile([]byte("\n"))
}

func (l *LogClient) WriteTestEntrySeperator() {
	entry := fmt.Sprintf("%s\n", strings.Repeat("=", 77))

	l.WriteFile([]byte(entry))
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
