package main

import "log"

type LogType int
const (
	Console LogType = 0
	File            = 1
)

type LogSeverity int
const (
	Debug LogSeverity = 0
	Info              = 1
	Warning           = 2
	Critical          = 3
)

type RatLogger struct {
	loggerType LogType
	filename   string
	LogChannel chan
	logLevel   LogSeverity
	
}

type LogPacket struct {
	Severity LogSeverity
	content  string
}
