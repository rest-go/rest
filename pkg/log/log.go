package log

import (
	"log"
	"os"
	"strings"
)

type Level int

const (
	InfoLevel Level = iota
	WarnLevel
	ErrorLevel
	DebugLevel
	TraceLevel
)

var (
	InfoLogger  = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime)
	WarnLogger  = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
	DebugLogger = log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime)
	TraceLogger = log.New(os.Stderr, "TRACE: ", log.Ldate|log.Ltime)

	logLevel Level
)

func init() { //nolint:gochecknoinits
	level := os.Getenv("REST_LOG_LEVEL")
	switch strings.ToLower(level) {
	case "trace":
		logLevel = TraceLevel
	case "error":
		logLevel = ErrorLevel
	case "warn":
		logLevel = WarnLevel
	case "info":
		logLevel = InfoLevel
	case "debug":
		logLevel = DebugLevel
	default:
		logLevel = ErrorLevel
	}
}

func SetLevel(level Level) {
	logLevel = level
}

func Info(v ...any) {
	InfoLogger.Print(v...)
}
func Infof(format string, v ...any) {
	InfoLogger.Printf(format, v...)
}

func Warn(v ...any) {
	if logLevel < WarnLevel {
		return
	}
	WarnLogger.Print(v...)
}
func Warnf(format string, v ...any) {
	if logLevel < WarnLevel {
		return
	}
	WarnLogger.Printf(format, v...)
}

func Error(v ...any) {
	if logLevel < ErrorLevel {
		return
	}
	ErrorLogger.Print(v...)
}
func Errorf(format string, v ...any) {
	if logLevel < ErrorLevel {
		return
	}
	ErrorLogger.Printf(format, v...)
}

func Debug(v ...any) {
	if logLevel < DebugLevel {
		return
	}
	DebugLogger.Print(v...)
}
func Debugf(format string, v ...any) {
	if logLevel < DebugLevel {
		return
	}
	DebugLogger.Printf(format, v...)
}

func Trace(v ...any) {
	if logLevel < TraceLevel {
		return
	}
	TraceLogger.Print(v...)
}
func Tracef(format string, v ...any) {
	if logLevel < TraceLevel {
		return
	}
	TraceLogger.Printf(format, v...)
}

func Fatal(v ...any) {
	log.Fatal(v...)
}
