package fdapm_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/foodora/go-ranger/fdapm"
	newrelic "github.com/newrelic/go-agent"
	"github.com/stretchr/testify/assert"
)

var newrelicApp newrelic.Application

func init() {
	config := newrelic.NewConfig("fdapm-newrelic-test", strings.Repeat(" ", 40))

	var err error

	newrelicApp, err = newrelic.NewApplication(config)
	if err != nil {
		panic(fmt.Errorf("Cannot create newrelic application: %s", err))
	}
}

func TestNewRelicMiddleware(t *testing.T) {
	newrelicMiddleware := fdapm.NewRelicMiddleware(newrelicApp)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		assert.Implements(t, (*newrelic.Transaction)(nil), w)
		called = true
	}

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	// call handler with middleware
	newrelicMiddleware(http.HandlerFunc(handler)).ServeHTTP(w, req)

	assert.True(t, called)
}

func TestNewRelicMiddleware_InjectTransaction(t *testing.T) {
	newrelicMiddleware := fdapm.NewRelicMiddleware(newrelicApp)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		txn := fdapm.NewRelicTransaction(req.Context())
		assert.NotNil(t, txn)
		assert.Equal(t, w, txn)
		called = true
	}

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	// call handler with middleware
	newrelicMiddleware(http.HandlerFunc(handler)).ServeHTTP(w, req)

	assert.True(t, called)
}

func TestNewRelicTransaction_WithoutUseMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		assert.Panics(t, func() {
			fdapm.NewRelicTransaction(req.Context())
		})
	}

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	handler(w, req)
}

func TestNewRelicStartSegment(t *testing.T) {
	newrelicMiddleware := fdapm.NewRelicMiddleware(newrelicApp)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		defer fdapm.NewRelicStartSegment(req.Context(), "my-segment").End()
		called = true
	}

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	// call handler with middleware
	newrelicMiddleware(http.HandlerFunc(handler)).ServeHTTP(w, req)

	assert.True(t, called)
}
