package fdapm

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	newrelic "github.com/newrelic/go-agent"
	"github.com/stretchr/testify/assert"
)

var newrelicApp newrelic.Application

func init() {
	config := newrelic.NewConfig("fdapm-newrelic-internal-test", strings.Repeat(" ", 40))

	var err error

	newrelicApp, err = newrelic.NewApplication(config)
	if err != nil {
		panic(fmt.Errorf("Cannot create newrelic application: %s", err))
	}
}

func TestNewRelicRoundTripper_DontCreateNewRelicOverNewRelicRoundTripper(t *testing.T) {
	txn := newrelicApp.StartTransaction("newrelic-test-transaction", nil, nil)
	defer txn.End()

	tr := &http.Transport{}
	c := &http.Client{
		Transport: tr,
	}

	WithNewRelic(c, txn)

	nrtr, _ := c.Transport.(*NewRelicRoundTripper)
	assert.Equal(t, tr, nrtr.orig)

	WithNewRelic(c, txn)

	nrtr2, _ := c.Transport.(*NewRelicRoundTripper)
	assert.NotEqual(t, nrtr, nrtr2)
	assert.Equal(t, tr, nrtr2.orig)
}
