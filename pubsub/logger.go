package pubsub

import (
	"log"
	"os"
)

// Logger is the interface used internally to log
type Logger interface {
	Printf(format string, v ...interface{})
}

// ErrorLogger logs with an error level
type ErrorLogger interface {
	Error(args ...interface{})
}

// DefaultLogger will be used as logger when create a new server using NewServer()
var DefaultLogger Logger

func init() {
	// Set default logger
	SetLogger(nil)
}

// SetLogger will be used as logger when create a new pub or sub instances using NewPublisher() and NewSubscriber(), but
func SetLogger(logger Logger) {
	if logger == nil {
		DefaultLogger = log.New(os.Stdout, "[pubsub] ", log.LstdFlags)
	} else {
		DefaultLogger = logger
	}
}
