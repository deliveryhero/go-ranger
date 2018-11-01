package fdapm_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foodora/go-ranger/fdapm/apmmock"

	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	"github.com/stretchr/testify/assert"

	"github.com/foodora/go-ranger/fdapm"
)

func TestNewRelicTransport(t *testing.T) {
	var mCalled bool

	expectedReq, _ := http.NewRequest(http.MethodGet, "http://www.foodora.de", nil)
	txn := apmmock.NewNRTransaction(t)
	m := fdapm.NewRelicTransport(txn)
	transport := m.Wrap(fdmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		mCalled = true
		assert.Equal(t, expectedReq, req)
		return nil, nil
	}))

	transport.RoundTrip(expectedReq)
	assert.True(t, mCalled)
	assert.True(t, txn.StartSegmentNowInvoked)
}

func TestNewRelicTransport_EmptyTransaction(t *testing.T) {
	var mCalled bool

	req, _ := http.NewRequest(http.MethodGet, "http://www.foodora.de", nil)
	m := fdapm.NewRelicTransport(nil)
	transport := m.Wrap(fdmiddleware.RoundTripperFunc(func(*http.Request) (*http.Response, error) {
		mCalled = true
		return nil, nil
	}))

	transport.RoundTrip(req)
	assert.True(t, mCalled)
}

func TestNewRelicClientMiddleware(t *testing.T) {
	var mCalled bool

	ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		mCalled = true
	}))
	defer ts.Close()

	c := &http.Client{
		Transport: http.DefaultTransport,
	}

	txn := apmmock.NewNRTransaction(t)
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	fdapm.NewRelicClientMiddleware(c, txn, req)

	assert.True(t, mCalled)
	assert.True(t, txn.StartSegmentNowInvoked)
	assert.Equal(t, http.DefaultTransport, c.Transport)
}
