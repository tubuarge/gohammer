package util

import (
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
