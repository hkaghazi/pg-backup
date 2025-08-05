package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	file   *os.File
	logger *log.Logger
}

func New(filename string) *Logger {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", 0)

	return &Logger{
		file:   file,
		logger: logger,
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.log("WARNING", format, args...)
}

func (l *Logger) log(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s: %s", timestamp, level, message)
}

func (l *Logger) Close() error {
	return l.file.Close()
}
