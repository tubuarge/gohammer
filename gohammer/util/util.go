package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	FileTimestampLayout   = "2006_01_02 15_04"
	LoggerTimestampLayout = "2006/01/02 15:04:03"
)

func StringInSlice(elem string, list []string) bool {
	for _, value := range list {
		if value == elem {
			return true
		}
	}
	return false
}

func ConvertStrToByte(strData string) []byte {
	return []byte(strData)
}

func GetFormattedTimestamp(layout string, timestamp time.Time) string {
	return timestamp.Format(layout)
}

func GetFormattedTimestampNow() string {
	now := time.Now()
	return now.Format(FileTimestampLayout)
}

func IsDirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateDir(path string) error {
	err := os.Mkdir(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func ParseDuration(str string) (time.Duration, error) {
	duration, err := time.ParseDuration(str)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func GetTestEntrySeperatorStr() string {
	return fmt.Sprintf("%s\n", strings.Repeat("=", 77))

}
