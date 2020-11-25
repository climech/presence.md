package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

const logFlags = log.Ldate | log.Ltime

// Logger wraps log.Logger, adding optional file-logging capability.
type Logger struct {
	*log.Logger
	file *os.File
}

func NewLogger() *Logger {
	return &Logger{Logger: log.New(os.Stdin, "", logFlags)}
}

func NewFileLogger(filename string, printToConsole bool) (*Logger, error) {
	dir, _ := filepath.Split(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	var writer io.Writer
	if printToConsole {
		writer = io.MultiWriter(os.Stdin, f)
	} else {
		writer = f
	}

	logger := &Logger{
		Logger: log.New(writer, "", logFlags),
		file:   f,
	}
	return logger, nil
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
