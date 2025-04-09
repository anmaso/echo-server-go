package logger

import (
	"log"
	"os"
	"sync"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	logger *log.Logger
	level  Level
	mu     sync.Mutex
}

// New creates a new Logger instance
func New(level Level) *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  level,
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.mu.Lock()
		l.logger.Printf("[DEBUG] "+format, v...)
		l.mu.Unlock()
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.mu.Lock()
		l.logger.Printf("[INFO] "+format, v...)
		l.mu.Unlock()
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.mu.Lock()
		l.logger.Printf("[WARN] "+format, v...)
		l.mu.Unlock()
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.mu.Lock()
		l.logger.Printf("[ERROR] "+format, v...)
		l.mu.Unlock()
	}
}
