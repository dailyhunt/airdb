package utils

import (
	"time"
)

const (
	EMPTY_STR = ""
)

func GetCurrentTime() string {
	t := time.Now()
	currentTime := t.Format("2006-01-02 15:04:05.000")
	return currentTime
}
