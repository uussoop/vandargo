// Package vandargo provides a secure integration with the Vandar payment gateway
// simple_logger.go implements a simple logger for testing purposes
package vandargo

import (
	"context"
	"fmt"
	"log"
	"os"
)

// SimpleLogger is a basic implementation of LoggerInterface for testing
type SimpleLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logLevel    string
}

// NewSimpleLogger creates a new simple logger with the specified log level
func NewSimpleLogger(level string) *SimpleLogger {
	return &SimpleLogger{
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime),
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
		logLevel:    level,
	}
}

// shouldLog determines whether the log should be output based on log level
func (l *SimpleLogger) shouldLog(level LogLevel) bool {
	switch l.logLevel {
	case "DEBUG":
		return true
	case "INFO":
		return level >= Info
	case "WARN":
		return level >= Warn
	case "ERROR":
		return level >= Error
	default:
		return level >= Info // Default to INFO level
	}
}

// Debug logs debug level messages
func (l *SimpleLogger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	if !l.shouldLog(Debug) {
		return
	}

	l.debugLogger.Printf("%s %v", message, fields)
}

// Info logs informational messages
func (l *SimpleLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	if !l.shouldLog(Info) {
		return
	}

	l.infoLogger.Printf("%s %v", message, fields)
}

// Warn logs warning messages
func (l *SimpleLogger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	if !l.shouldLog(Warn) {
		return
	}

	l.warnLogger.Printf("%s %v", message, fields)
}

// Error logs error messages
func (l *SimpleLogger) Error(ctx context.Context, message string, err error, fields map[string]interface{}) {
	if !l.shouldLog(Error) {
		return
	}

	errMsg := ""
	if err != nil {
		errMsg = fmt.Sprintf(" Error: %v", err)
	}

	l.errorLogger.Printf("%s%s %v", message, errMsg, fields)
}
