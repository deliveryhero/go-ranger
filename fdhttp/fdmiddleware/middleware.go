package fdmiddleware

import "net/http"

// MiddlewareFunc is a easy way to convert a function to a interface Doer
type MiddlewareFunc func(next http.Handler) http.Handler

// Do will be called for each middleware until http.Client
func (f MiddlewareFunc) Wrap(next http.Handler) http.Handler {
	return f(next)
}

// Middleware specify a interface to http calls
type Middleware interface {
	Wrap(next http.Handler) http.Handler
}
