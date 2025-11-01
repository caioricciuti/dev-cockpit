package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	instance *Logger
	once     sync.Once
)

// Logger handles all application logging
type Logger struct {
	file   *os.File
	logger *log.Logger
	level  LogLevel
	mu     sync.Mutex
	path   string
}

// Initialize sets up the logger singleton
func Initialize(debugMode bool) error {
	var err error
	once.Do(func() {
		instance = &Logger{
			level: INFO,
		}
		if debugMode {
			instance.level = DEBUG
		}
		// Always try to set up logging, but gracefully fall back on failure
		if e := instance.setupLogFile(); e != nil {
			// Fallback to temp dir
			_ = instance.setupFallbackLogger()
			err = nil
		}
	})
	return err
}

// GetLogger returns the logger instance
func GetLogger() *Logger {
	if instance == nil {
		Initialize(false)
	}
	return instance
}

func (l *Logger) setupLogFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(homeDir, ".devcockpit")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, "debug.log")

	// Open or create log file
	l.file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.logger = log.New(l.file, "", 0)
	l.path = logPath

	// Log initialization
	l.Info("=== Dev Cockpit Started ===")
	l.Info("Log file: %s", logPath)
	l.Info("Debug mode: %v", l.level == DEBUG)

	return nil
}

// setupFallbackLogger configures logging to a safe fallback location or stderr
func (l *Logger) setupFallbackLogger() error {
	// First try temp directory
	tmpPath := filepath.Join(os.TempDir(), "devcockpit-debug.log")
	if f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		l.file = f
		l.logger = log.New(f, "", 0)
		l.path = tmpPath
		l.Info("Using fallback log path: %s", tmpPath)
		l.Info("Debug mode: %v", l.level == DEBUG)
		return nil
	}

	// As last resort, log to stderr only
	l.file = nil
	l.logger = log.New(os.Stderr, "", 0)
	l.path = ""
	l.Info("Falling back to stderr logging (no file)")
	return nil
}

// Close closes the log file
func (l *Logger) Close() {
	if l.file != nil {
		l.Info("=== Dev Cockpit Stopped ===")
		l.file.Close()
	}
}

// GetLogPath returns the path to the log file
func GetLogPath() string {
	if instance != nil && instance.path != "" {
		return instance.path
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".devcockpit", "debug.log")
}

// log writes a formatted log message
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l == nil || l.logger == nil || level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get caller information
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)

	levelStr := ""
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	logLine := fmt.Sprintf("%s [%s] %s:%d - %s", timestamp, levelStr, file, line, message)
	l.logger.Println(logLine)

	// Also print to stdout if in debug mode
	if l.level == DEBUG {
		fmt.Fprintln(os.Stderr, logLine)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// LogCommand logs a command execution
func (l *Logger) LogCommand(cmd string, args []string, output []byte, err error) {
	fullCmd := fmt.Sprintf("%s %v", cmd, args)
	l.Debug("Executing command: %s", fullCmd)

	if len(output) > 0 {
		l.Debug("Command output: %s", string(output))
	}

	if err != nil {
		l.Error("Command failed: %s, error: %v", fullCmd, err)
	} else {
		l.Debug("Command succeeded: %s", fullCmd)
	}
}

// Static functions for easier access
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func LogCommand(cmd string, args []string, output []byte, err error) {
	GetLogger().LogCommand(cmd, args, output, err)
}
