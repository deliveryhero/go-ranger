package fdhttp

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// NewPanicHandler return a handler, used by default, to deal
// when a panic happened inside of your handler
func NewPanicHandler() func(http.ResponseWriter, *http.Request, interface{}) {
	return panicHandler
}

func panicHandler(w http.ResponseWriter, req *http.Request, rcv interface{}) {
	// log stack trace
	defaultLogger.Printf("%s", debug.Stack())

	ResponseJSON(w, http.StatusInternalServerError, &Error{
		Code:    "panic",
		Message: fmt.Sprint(rcv),
	})
}
