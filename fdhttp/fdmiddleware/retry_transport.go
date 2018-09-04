package fdmiddleware

import (
	"net/http"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
)

type RetryTransport struct {
	maxRetries  int
	backoffFunc fdbackoff.Func
}

// NewRetryTransport will retry maxRetries using backoffFunc to wait between
// these calls. Once we have a successful call, status code less than 500 we'll stop.
// Status code 429 - Too Many Request will trigger a retry as well.
func NewRetryTransport(maxRetries int, backoffFunc fdbackoff.Func) *RetryTransport {
	return &RetryTransport{
		maxRetries:  maxRetries,
		backoffFunc: backoffFunc,
	}
}

func (m *RetryTransport) Wrap(next http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
		for retry := 0; retry < m.maxRetries; retry++ {
			resp, err = next.RoundTrip(req)
			if err == nil && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
				// we can consider this situation as a successful call, let's return
				return
			}

			time.Sleep(m.backoffFunc(retry + 1))
		}

		return
	})
}
