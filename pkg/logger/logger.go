package logger

import (
	"log"
	"os"
)

// Logger interface defines the logging operations
type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type logger struct {
	log *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) Logger {
	return &logger{
		log: log.New(os.Stdout, prefix, log.LstdFlags),
	}
}

func (l *logger) Info(args ...interface{}) {
	l.log.Println(args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.log.Printf(format, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.log.Println(args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.log.Printf(format, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.log.Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.log.Fatalf(format, args...)
}
