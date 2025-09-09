package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// ANSI colors
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var currentLevel = INFO
var showTimestamp = true

func SetLevel(level Level) {
	currentLevel = level
}

func ShowTimestamp(enabled bool) {
	showTimestamp = enabled
}
func logWithColor(level Level, color string, tag string, format string, v ...interface{}) {
	if level < currentLevel {
		return
	}
	msg := fmt.Sprintf(format, v...)

	prefix := ""
	if showTimestamp {
		prefix = time.Now().Format("15:04:05") + " "
	}
	log.Printf("%s%s[%s]%s %s", prefix, color, tag, reset, msg)
}

func Debug(format string, v ...interface{}) {
	logWithColor(DEBUG, cyan, "DEBUG", format, v...)
}

func Info(format string, v ...interface{}) {
	logWithColor(INFO, green, "INFO", format, v...)
}

func Warn(format string, v ...interface{}) {
	logWithColor(WARN, yellow, "WARN", format, v...)
}

func Error(format string, v ...interface{}) {
	logWithColor(ERROR, red, "ERROR", format, v...)
	os.Exit(1)
}
