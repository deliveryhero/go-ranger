package fdmiddleware

import "net/http"

//// Client Transport

// RoundTripperFunc is a easy way to convert a function to a interface http.RoundTripper
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip will be called for each middleware until http.Client
func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// ClientMiddlewareFunc will wrap all Do calls
type ClientMiddlewareFunc func(next http.RoundTripper) http.RoundTripper

// Wrap will be called for each middleware until http.Client
func (f ClientMiddlewareFunc) Wrap(next http.RoundTripper) http.RoundTripper {
	return f(next)
}

// ClientMiddleware specify a interface to http calls
type ClientMiddleware interface {
	Wrap(next http.RoundTripper) http.RoundTripper
}
