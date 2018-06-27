package fdhttp

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
)

// A type that satisfies fdhttp.Handler can be registered as a handler on fdhttp.Router.
type Handler interface {
	// This method will be called right before your fdhttp.Server run or when router.Init()
	// is called and it needes to register all endpoints that your handler implements.
	Init(*Router)
}

// EndpointFunc is the method signature to deal with http requests.
//
// See Also
//
// Router.GET(), Router.POST(), Router.PUT(), Router.DELETE()
// or functions that are compatible with standard library
// Router.StdGET(), Router.StdPOST(), Router.StdPUT(), Router.StdDELETE()
type EndpointFunc func(context.Context) (int, interface{})

// JSONer if your response (second return parameter of fdhttp.EndpointFunc)
// implements this interface it'll be called to send as response.
type JSONer interface {
	JSON() interface{}
}

func methodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	ResponseJSON(w, http.StatusMethodNotAllowed, &Error{
		Code:    "method_not_allowed",
		Message: fmt.Sprintf("Method '%s' is not allowed to access '%s'", req.Method, req.URL.String()),
	})
}

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	ResponseJSON(w, http.StatusNotFound, &Error{
		Code:    "not_found",
		Message: fmt.Sprintf("URL '%s' was not found", req.URL.String()),
	})
}

func panicHandler(w http.ResponseWriter, req *http.Request, rcv interface{}) {
	// log stack trace
	defaultLogger.Printf("%s", debug.Stack())

	ResponseJSON(w, http.StatusInternalServerError, &Error{
		Code:    "panic",
		Message: fmt.Sprint(rcv),
	})
}
