package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel defines the severity of log messages
type LogLevel string

const (
	// Log levels
	LogDebug   LogLevel = "DEBUG"
	LogInfo    LogLevel = "INFO"
	LogWarning LogLevel = "WARNING"
	LogError   LogLevel = "ERROR"
	LogSuccess LogLevel = "SUCCESS"
)

// Logger is the main struct for logging operations
type Logger struct {
	file   *os.File
	writer io.Writer
}

var logInstance *Logger

// InitLogger initializes the logger with the specified log file
func InitLogger(logFilePath string) (*Logger, error) {
	if logInstance != nil {
		return logInstance, nil
	}

	// Create logs directory if it doesn't exist
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file (create if not exists, append if exists)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// In development mode, write to both console and file
	// In production mode, write only to file
	var writer io.Writer
	if gin.Mode() == gin.DebugMode {
		writer = io.MultiWriter(os.Stdout, file)
	} else {
		writer = file
	}

	logInstance = &Logger{
		file:   file,
		writer: writer,
	}

	return logInstance, nil
}

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	if logInstance == nil {
		// Default to a logs/app.log file if not initialized
		logger, err := InitLogger("logs/app.log")
		if err != nil {
			// Fall back to stdout if file logging fails
			return &Logger{writer: os.Stdout}
		}
		return logger
	}
	return logInstance
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// formatMessage formats a log message with timestamp, level, and caller info
func (l *Logger) formatMessage(level LogLevel, message string) string {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	callerInfo := "unknown"
	if ok {
		parts := strings.Split(file, "/")
		if len(parts) >= 2 {
			callerInfo = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Return formatted log message
	return fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, level, callerInfo, message)
}

// Log logs a message with the specified level
func (l *Logger) Log(level LogLevel, message string) {
	formattedMessage := l.formatMessage(level, message)
	fmt.Fprint(l.writer, formattedMessage)
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	if gin.Mode() == gin.DebugMode {
		l.Log(LogDebug, message)
	}
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.Log(LogInfo, message)
}

// Warning logs a warning message
func (l *Logger) Warning(message string) {
	l.Log(LogWarning, message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.Log(LogError, message)
}

// Success logs a success message
func (l *Logger) Success(message string) {
	l.Log(LogSuccess, message)
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(c *gin.Context, latency time.Duration) {
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}

	clientIP := c.ClientIP()
	method := c.Request.Method
	statusCode := c.Writer.Status()
	userAgent := c.Request.UserAgent()

	// Format message
	message := fmt.Sprintf("%s | %3d | %13v | %15s | %s | %s",
		method, statusCode, latency, clientIP, path, userAgent)

	// Choose log level based on status code
	switch {
	case statusCode >= 500:
		l.Error(message)
	case statusCode >= 400:
		l.Warning(message)
	default:
		l.Info(message)
	}
}
