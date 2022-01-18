package main

import (
	"log"
	"os"
	"fmt"
)

type LogType uint8
const (
	Console LogType = 1 << iota
	File //Could be any IO writer. Basically just uses a non-default logger. But during new it assumes it's a file.
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
	severity LogSeverity
	content  string
}

func NewLogger(logtype LogType, filename string, logLevel LogSeverity) *RatLogger {
	var l RatLogger
	l.loggerType = logtype
	l.filename = filename
	l.logLevel = logLevel
	
	if l.loggerType&File > 0 {
		f, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		l.logger = *log.New(f, "", log.Ldate | log.Ltime)
	}
	
	return &l
}

func (l RatLogger) HandleLogs() {
	for m := range l.LogChannel {
		if m.severity >= l.logLevel {
			if l.loggerType&Console > 0 {
				log.Println(fmt.Sprintf("%s: %s", sevToStr(m.severity), m.content))
			}
			if l.loggerType&File > 0 {
				l.logger.Println(fmt.Sprintf("%s: %s", sevToStr(m.severity), m.content))
			}
		}
	}
}

func (l RatLogger) Log(sev LogSeverity, m string) {
	l.LogChannel <- LogPacket {sev, m}
}

func sevToStr(s LogSeverity) string {
	switch s {
		case Debug:
			return "DEBUG"
		case Info:
			return "INFO"
		case Warning:
			return "WARNING"
		case Critical:
			return "CRITICAL"
	}
	return "UNKNOWN"
}

func (l *RatLogger) Close() {
	if l.loggerType&File > 0 {
		l.file.Close()
	}
	close(l.LogChannel)
}




