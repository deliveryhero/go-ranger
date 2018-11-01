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
