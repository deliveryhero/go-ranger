package fdhttp

import "net/http"

// ClientMiddleware will wrap all Do calls
type ClientMiddleware func(next Doer) Doer

// DoerFunc is a easy way to convert a function to a interface Doer
type DoerFunc func(req *http.Request) (*http.Response, error)

// Do will be called for each middleware until http.Client
func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Doer specify a interface to http calls
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}
