package xormigrate

import (
	"io"
	"log"
	"os"
)

type LoggerInterface interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
}

var logger LoggerInterface = defaultLogger()

// SetLogger sets the Xormigrate logger.
func (x *Xormigrate) SetLogger(l LoggerInterface) {
	logger = l
}

func defaultLogger() *XormigrateLogger {
	return &XormigrateLogger{log.New(os.Stdout, "[xormigrate] ", 0)}
}

// DefaultLogger sets a Xormigrate logger with default settings
// e.g. "[xormigrate] message".
func (x *Xormigrate) DefaultLogger() {
	x.SetLogger(defaultLogger())
}

// NilLogger sets a Xormigrate logger that discards all messages.
func (x *Xormigrate) NilLogger() {
	x.SetLogger(&XormigrateLogger{log.New(io.Discard, "", 0)})
}

// NewLogger sets a Xormigrate logger with a specified io.Writer.
func (x *Xormigrate) NewLogger(writer io.Writer) {
	x.SetLogger(&XormigrateLogger{log.New(writer, "", 0)})
}

type XormigrateLogger struct {
	*log.Logger
}

// Debug prints a Debug message.
func (l *XormigrateLogger) Debug(v ...any) {
	l.Print(v...)
}

// Debugf prints a formatted Debug message.
func (l *XormigrateLogger) Debugf(format string, v ...any) {
	l.Printf(format, v...)
}

// Info prints an Info message.
func (l *XormigrateLogger) Info(v ...any) {
	l.Print(v...)
}

// Infof prints a formatted Info message.
func (l *XormigrateLogger) Infof(format string, v ...any) {
	l.Printf(format, v...)
}

// Warn prints a Warning message.
func (l *XormigrateLogger) Warn(v ...any) {
	l.Print(v...)
}

// Warnf prints a formatted Warning message.
func (l *XormigrateLogger) Warnf(format string, v ...any) {
	l.Printf(format, v...)
}

// Error prints an Error message.
func (l *XormigrateLogger) Error(v ...any) {
	l.Print(v...)
}

// Errorf prints a formatted Error message.
func (l *XormigrateLogger) Errorf(format string, v ...any) {
	l.Printf(format, v...)
}
