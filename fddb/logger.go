package fddb

import (
	"log"
	"os"
)

// Logger is the interface used internally to log
type Logger interface {
	Printf(format string, v ...interface{})
}

// defaultLogger will be used as logger when create a new server using NewServer()
var defaultLogger Logger

func init() {
	// Set default logger
	SetLogger(nil)
}

// SetLogger will be used as logger when create a new server using NewServer(), but
// it's possible update individualy a server.Logger later.
func SetLogger(logger Logger) {
	if logger == nil {
		defaultLogger = log.New(os.Stdout, "[fddb] ", log.LstdFlags)
	} else {
		defaultLogger = logger
	}
}
