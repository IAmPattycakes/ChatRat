package main

import (
	"log"
	"os"
)

type LogType uint8
const (
	Console LogType = 1 << iota
	File
	
)

type LogSeverity int
const (
	Debug LogSeverity = 0
	Info              = 1
	Warning           = 2
	Critical          = 3
)

//RatLogger is a wrapper around the normal Go logger that just adds logging levels 
type RatLogger struct {
	loggerType  LogType
	filename    string
	LogChannel  chan LogPacket
	logLevel    LogSeverity
	logger      log.Logger
	file        os.File
}

type LogPacket struct {
	Severity LogSeverity
	content  string
}

func NewLogger(logtype LogType, filename string, logLevel LogSeverity) *RatLogger {
	var l RatLogger
	l.loggerType = logtype
	l.filename = filename
	l.logLevel = logLevel
	
	f, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("This is a test log entry")
	return &l
}

func (l *RatLogger) HandleLogs() {
	
}

func (l *RatLogger) Close() {
	defer l.file.Close()
	close(l.LogChannel)
}




