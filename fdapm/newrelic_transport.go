package fdapm

import (
	"net/http"
	"strconv"
	"time"

	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"

	newrelic "github.com/newrelic/go-agent"
)

// NewRelicTransport return a fdmiddleware.ClientMiddleware to instrument with NewRelic
// your http calls.
// The basic way to use is:
// 		req := http.NewRequest(http.MethodGet, "http://www.foodora.de", nil)
//
// 		// txn can be a transaction that you create by yourself, or
// 		// if you're using fdapm.NewRelicMiddleware in your http server it'll be your
// 		// w http.ResponseWriter (that you receive in your handler).
//		// Also if you implement your handler as fdhttp.EndpointFunc (receiving only ctx context.Context)
//		// you can get it like this:
// 		txn := fdapm.NewRelicTransaction(ctx)
//  	httpClient := &http.Client{}
// 		httpClient.Transport = NewRelicTransport(txn).Wrap(httpClient.Transport)
// 		resp, err := httpClient.Do(req)
//
// If you have a global http.Client that you're reusing between different request,
// see fdapm.NewRelicClientMiddleware.
func NewRelicTransport(txn newrelic.Transaction) fdmiddleware.ClientMiddleware {
	return fdmiddleware.ClientMiddlewareFunc(func(next http.RoundTripper) http.RoundTripper {
		return fdmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if txn == nil {
				return next.RoundTrip(req)
			}

			// Add header to report request queueing
			req.Header.Set("X-Request-Start", strconv.FormatInt(time.Now().UnixNano()/int64(time.Microsecond), 10))

			seg := newrelic.StartExternalSegment(txn, req)
			resp, err := next.RoundTrip(req)
			seg.Response = resp
			seg.End()

			return resp, err
		})
	})
}

// NewRelicClientMiddleware makes the http call instrumenting with newrelic.Transaction.
// It's common that you have only one http.Client and you share it across multiple requests or
// different parts of your application. The problem is if you're calling a external service
// for each request you cannot add the transport using fdapm.NewRelicTransport to your
// http client otherwise you're reusing the same transaction for different calls.
// This method use the http.Client that you provided and add the middleware on it, only to execute
// this call. Doing that, the transport is used only in a single request, and next calls will not
// have the middleware anymore.
func NewRelicClientMiddleware(httpClient *http.Client, txn newrelic.Transaction, req *http.Request) (*http.Response, error) {
	c := &http.Client{}
	*c = *httpClient
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}

	c.Transport = NewRelicTransport(txn).Wrap(c.Transport)
	return c.Do(req)
}
