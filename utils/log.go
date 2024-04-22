package utils

import (
	"fmt"
	"log"
)

type LogLevel uint8

const (
	LogLevelNothing LogLevel = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

// LogLevel of the runtime; if need to change it, modify it here.
const ExecLogLevel LogLevel = LogLevelDebug

func ErrorLog(message string, args ...interface{}) {
	if ExecLogLevel >= LogLevelError {
		log.Printf(message, args...)
	}
}

func InfoLog(message string, args ...interface{}) {
	if ExecLogLevel >= LogLevelInfo {
		log.Printf(message, args...)
	}
}

func DebugLog(message string, args ...interface{}) {
	if ExecLogLevel >= LogLevelDebug {
		log.Printf(message, args...)
	}
}

func ErrorPrintf(message string, args ...interface{}) {
	if ExecLogLevel >= LogLevelError {
		fmt.Printf(message, args...)
	}
}

func InfoPrintf(message string, args ...interface{}) {
	if ExecLogLevel >= LogLevelInfo {
		fmt.Printf(message, args...)
	}
}

func DebugPrintf(message string, args ...interface{}) {
	if ExecLogLevel >= LogLevelDebug {
		fmt.Printf(message, args...)
	}
}
