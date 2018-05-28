package fdapm

import (
	"net/http"

	"github.com/foodora/go-ranger/fdhttp"
	newrelic "github.com/newrelic/go-agent"
)

// NewRelicClientMiddleware call fdhttp.Doer instrumenting with NewRelic
// The basic way to use this http client middleware is:
// 		req := http.NewRequest(http.MethodGet, "www.google.com", nil)
//
// 		// txn can be a transaction that you create by yourself, or
// 		// if you're using fdapm.NewRelicMiddleware in your server it'll be your
// 		// w http.ResponseWriter (that you receive in your handler)
//		// but if you implemented your handler receiving only ctx context.Context
//		// you can get it like this:
// 		// txn := NewRelicTransaction(ctx)
// 		resp, err := fdapm.NewRelicClientMiddleware(http.DefaultClient, txn, req)
func NewRelicClientMiddleware(c fdhttp.Doer, txn newrelic.Transaction, req *http.Request) (*http.Response, error) {
	if txn == nil {
		return c.Do(req)
	}

	s := newrelic.StartExternalSegment(txn, req)
	resp, err := c.Do(req)
	s.Response = resp
	s.End()

	return resp, err
}
