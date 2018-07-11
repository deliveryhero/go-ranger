package fdhttp_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func newClientMiddleware(called *bool) fdhttp.ClientMiddleware {
	return func(next fdhttp.Doer) fdhttp.Doer {
		return fdhttp.DoerFunc(func(req *http.Request) (*http.Response, error) {
			*called = true
			return next.Do(req)
		})
	}
}

func TestClient_CallMiddleware(t *testing.T) {
	var called bool
	m := newClientMiddleware(&called)

	c := fdhttp.NewClient()
	c.Use(m)

	var srvCalled bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srvCalled = true
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBufferString(""))
	resp, err := c.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.True(t, called)
	assert.True(t, srvCalled)
}

func TestClientRetryMiddleware(t *testing.T) {
	middleware := fdhttp.NewClientRetryMiddleware(4, fdbackoff.Fixed(time.Millisecond))

	c := fdhttp.NewClient()
	c.Use(middleware)

	var srvCalled int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		srvCalled++

		var status int
		switch srvCalled {
		case 1:
			status = http.StatusInternalServerError
		case 2:
			status = http.StatusTooManyRequests
		case 3:
			status = http.StatusOK
		default:
			t.Error(t, "It was not expected this call")
			status = http.StatusInternalServerError
		}

		w.WriteHeader(status)
		fmt.Fprint(w, http.StatusText(status))
	}))
	defer ts.Close()

	resp, err := c.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, 3, srvCalled)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "OK", string(body))
}
