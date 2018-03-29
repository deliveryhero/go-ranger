package fdhttp

import (
	"io/ioutil"
	"log"
	"sync"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

var DefaultLogger = log.New(ioutil.Discard, "", 0)

func Un(f func()) {
	f()
}

func Lock(x sync.Locker) func() {
	x.Lock()
	return func() { x.Unlock() }
}
