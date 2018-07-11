package fdhttp

import (
	"net/http"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
)

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

// NewClientRetryMiddleware will retry maxRetries using backoffFunc to wait between
// these calls. Once we have a successful call, status code less than 500 we'll stop.
// Status code 429 - Too Many Request will trigger a retry as well.
func NewClientRetryMiddleware(maxRetries int, backoffFunc fdbackoff.Func) ClientMiddleware {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (resp *http.Response, err error) {
			for retry := 0; retry < maxRetries; retry++ {
				resp, err = next.Do(req)
				if err == nil && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
					// we can consider this situation as a successful call, let's return
					return
				}

				time.Sleep(backoffFunc(retry + 1))
			}

			return
		})
	}
}
