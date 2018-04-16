package fdapm

import (
	"net/http"

	newrelic "github.com/newrelic/go-agent"
)

type NewRelicRoundTripper struct {
	http.RoundTripper
	orig http.RoundTripper
}

func WithNewRelic(c *http.Client, txn newrelic.Transaction) *http.Client {
	if c == nil {
		c = &http.Client{}
	}

	if transport, ok := c.Transport.(*NewRelicRoundTripper); ok {
		// our transport is already a new relic one, let's create a new one removing the current one
		c.Transport = &NewRelicRoundTripper{
			RoundTripper: newrelic.NewRoundTripper(txn, transport.orig),
			orig:         transport.orig,
		}
	} else {
		c.Transport = &NewRelicRoundTripper{
			RoundTripper: newrelic.NewRoundTripper(txn, c.Transport),
			orig:         c.Transport,
		}
	}

	return c
}
