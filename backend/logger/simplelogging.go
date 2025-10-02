package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

// CustomLogger provides logging with enhanced details.
type CustomLogger struct {
	logger *log.Logger
}

// NewCustomLogger creates a new CustomLogger instance.
func NewCustomLogger(output *os.File, prefix string, flags int) *CustomLogger {
	return &CustomLogger{
		logger: log.New(output, prefix, flags),
	}
}

// Info logs an informational message with caller details.
func (l *CustomLogger) Info(message string) {
	l.logWithCaller("INFO", message)
}

// Warn logs a warning message with caller details.
func (l *CustomLogger) Warn(message string) {
	l.logWithCaller("WARN", message)
}

// Error logs an error message with caller details.
func (l *CustomLogger) Error(message string) {
	l.logWithCaller("ERROR", message)
}

// logWithCaller retrieves caller information and logs the message.
func (l *CustomLogger) logWithCaller(level, message string) {
	_, file, line, ok := runtime.Caller(3) // Caller(3) gets the caller of the Info/Error method
	if !ok {
		file = "???"
		line = 0
	}
	funcName := "???"
	pc, _, _, ok := runtime.Caller(1) // Caller(1) gets the immediate caller of logWithCaller
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = fn.Name()
		}
	}

	formattedMessage := fmt.Sprintf("%s %s:%d %s() - %s", level, path.Base(file), line, path.Base(funcName), message)
	l.logger.Println(formattedMessage)
}

// LogInfo is a convenience function to log informational messages with context.
func LogInfo(context, function, message string) {
	logger := NewCustomLogger(os.Stdout, "", log.LstdFlags)
	logger.Info(fmt.Sprintf("%s - %s: %s", context, function, message))
}

// LogWarn is a convenience function to log warning messages with context.
func LogWarn(context, function, message string) {
	logger := NewCustomLogger(os.Stdout, "", log.LstdFlags)
	logger.Warn(fmt.Sprintf("%s - %s: %s", context, function, message))
}

// LogError is a convenience function to log errors with context.
func LogError(context, function string, err error) {
	logger := NewCustomLogger(os.Stdout, "", log.LstdFlags)
	logger.Error(fmt.Sprintf("%s - %s: %v", context, function, err))
}
