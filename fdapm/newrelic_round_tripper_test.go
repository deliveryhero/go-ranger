package fdapm_test

import (
	"net/http"
	"testing"

	"github.com/foodora/go-ranger/fdapm"
	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func TestNewRelicRoundTripper_WithoutClient(t *testing.T) {
	txn := newrelicApp.StartTransaction("newrelic-test-transaction", nil, nil)
	defer txn.End()

	c := fdapm.WithNewRelic(nil, txn)
	assert.IsType(t, &fdapm.NewRelicRoundTripper{}, c.Transport)
}

func TestNewRelicRoundTripper_WithStdClient(t *testing.T) {
	txn := newrelicApp.StartTransaction("newrelic-test-transaction", nil, nil)
	defer txn.End()

	c := &http.Client{}
	fdapm.WithNewRelic(c, txn)
	assert.IsType(t, &fdapm.NewRelicRoundTripper{}, c.Transport)
}

func TestNewRelicRoundTripper_WithFDHTTTPClient(t *testing.T) {
	txn := newrelicApp.StartTransaction("newrelic-test-transaction", nil, nil)
	defer txn.End()

	c := fdhttp.NewClient()
	fdapm.WithNewRelic(&c.Client, txn)
	assert.IsType(t, &fdapm.NewRelicRoundTripper{}, c.Transport)

	c.Get("http://www.foodora.net")
}
