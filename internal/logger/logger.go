package logger

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const MaxLogLines = 1000

// LogCallback is called when a new log entry is created
type LogCallback func(LogEntry)

// Logger wraps logrus and keeps logs in memory
type Logger struct {
	*logrus.Logger
	buffer    []LogEntry
	bufferMu  sync.RWMutex
	maxLines  int
	callback  LogCallback
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// New creates a new logger
func New(level string) (*Logger, error) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	l.SetLevel(logLevel)

	logger := &Logger{
		Logger:   l,
		buffer:   make([]LogEntry, 0, MaxLogLines),
		maxLines: MaxLogLines,
	}

	// Add hook to capture logs
	l.AddHook(logger)

	return logger, nil
}

// Levels implements logrus.Hook interface
func (l *Logger) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire implements logrus.Hook interface
func (l *Logger) Fire(entry *logrus.Entry) error {
	logEntry := LogEntry{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Message:   entry.Message,
	}

	l.bufferMu.Lock()
	l.buffer = append(l.buffer, logEntry)

	// Keep only last maxLines entries
	if len(l.buffer) > l.maxLines {
		l.buffer = l.buffer[len(l.buffer)-l.maxLines:]
	}

	// Get callback before unlock
	callback := l.callback
	l.bufferMu.Unlock()

	// Call callback outside lock to avoid deadlock
	if callback != nil {
		callback(logEntry)
	}

	return nil
}

// GetLogs returns all logs in memory
func (l *Logger) GetLogs() []LogEntry {
	l.bufferMu.RLock()
	defer l.bufferMu.RUnlock()

	// Return a copy
	logs := make([]LogEntry, len(l.buffer))
	copy(logs, l.buffer)
	return logs
}

// SetCallback sets a callback to be called when a new log entry is created
func (l *Logger) SetCallback(callback LogCallback) {
	l.bufferMu.Lock()
	defer l.bufferMu.Unlock()
	l.callback = callback
}

// Clear clears the log buffer
func (l *Logger) Clear() {
	l.bufferMu.Lock()
	defer l.bufferMu.Unlock()
	l.buffer = make([]LogEntry, 0, l.maxLines)
}
