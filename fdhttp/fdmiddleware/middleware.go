package fdmiddleware

import "net/http"

//// Server Middleware

// MiddlewareFunc is a easy way to convert a function to a interface Middleware
type MiddlewareFunc func(next http.Handler) http.Handler

// Wrap will be called for each middleware until the main handler
func (f MiddlewareFunc) Wrap(next http.Handler) http.Handler {
	return f(next)
}

// Middleware specify a interface to http calls
type Middleware interface {
	Wrap(next http.Handler) http.Handler
}

//// Client Middleware

// Doer specify a interface to http calls
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoerFunc is a easy way to convert a function to a interface Doer
type DoerFunc func(req *http.Request) (*http.Response, error)

// Do will be called for each middleware until http.Client
func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// ClientMiddleware will wrap all Do calls
type ClientMiddlewareFunc func(next Doer) Doer

// Do will be called for each middleware until http.Client
func (f ClientMiddlewareFunc) Wrap(next Doer) Doer {
	return f(next)
}

// Middleware specify a interface to http calls
type ClientMiddleware interface {
	Wrap(next Doer) Doer
}
