package fdapm

import (
	"net/http"

	"github.com/foodora/go-ranger/fdhttp"
	newrelic "github.com/newrelic/go-agent"
)

// NewRelicClientMiddleware call fdhttp.Doer instrumenting with NewRelic
func NewRelicClientMiddleware(c fdhttp.Doer, txn newrelic.Transaction, req *http.Request) (*http.Response, error) {
	s := newrelic.StartExternalSegment(txn, req)
	resp, err := c.Do(req)
	s.Response = resp
	s.End()
	return resp, err
}
