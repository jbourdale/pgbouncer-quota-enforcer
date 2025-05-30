package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger defines the interface for application logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	WithField(key string, value interface{}) Logger
}

// SimpleLogger implements a basic logger
type SimpleLogger struct {
	logger *log.Logger
	fields map[string]interface{}
}

// NewSimpleLogger creates a new SimpleLogger instance
func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, "", 0),
		fields: make(map[string]interface{}),
	}
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	l.logWithLevel("INFO", msg, args...)
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	l.logWithLevel("ERROR", msg, args...)
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	l.logWithLevel("DEBUG", msg, args...)
}

// WithField returns a new logger with an additional field
func (l *SimpleLogger) WithField(key string, value interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &SimpleLogger{
		logger: l.logger,
		fields: newFields,
	}
}

func (l *SimpleLogger) logWithLevel(level, msg string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	formattedMsg := fmt.Sprintf(msg, args...)

	// Add fields to the message
	fieldsStr := ""
	if len(l.fields) > 0 {
		fieldsStr = " ["
		first := true
		for k, v := range l.fields {
			if !first {
				fieldsStr += ", "
			}
			fieldsStr += fmt.Sprintf("%s=%v", k, v)
			first = false
		}
		fieldsStr += "]"
	}

	logLine := fmt.Sprintf("[%s] %s: %s%s", timestamp, level, formattedMsg, fieldsStr)
	l.logger.Println(logLine)
}
