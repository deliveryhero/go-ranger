package fdhttp_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	"github.com/stretchr/testify/assert"
)

func newClientMiddleware(called *bool) fdmiddleware.ClientMiddleware {
	return fdmiddleware.ClientMiddlewareFunc(func(next http.RoundTripper) http.RoundTripper {
		return fdmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			*called = true
			return next.RoundTrip(req)
		})
	})
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

type testConn struct {
	net.Conn
	closeFn func() error
}

func (c *testConn) Close() error {
	return c.closeFn()
}

func middlewareConnCount(activeConns *int32) fdmiddleware.ClientMiddleware {
	return fdmiddleware.ClientMiddlewareFunc(func(next http.RoundTripper) http.RoundTripper {
		nextTr := next.(*http.Transport)
		tr := &http.Transport{
			MaxIdleConns:          nextTr.MaxIdleConns,
			MaxIdleConnsPerHost:   nextTr.MaxIdleConnsPerHost,
			IdleConnTimeout:       nextTr.IdleConnTimeout,
			TLSHandshakeTimeout:   nextTr.TLSHandshakeTimeout,
			ExpectContinueTimeout: nextTr.ExpectContinueTimeout,
		}

		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			atomic.AddInt32(activeConns, 1)

			conn, _ := next.(*http.Transport).DialContext(ctx, network, addr)
			tConn := &testConn{
				Conn: conn,
				closeFn: func() error {
					atomic.AddInt32(activeConns, -1)
					return conn.Close()
				},
			}
			return tConn, nil
		}
		return tr
	})
}

func httpGetParallel(t *testing.T, c fdhttp.Client, url string, times int) {
	var wg sync.WaitGroup
	for i := 0; i < times; i++ {
		wg.Add(1)
		go func() {
			resp, err := c.Get(url)
			assert.NoError(t, err)
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestClient_MaxIdleConns(t *testing.T) {
	tests := []int{1, 2, 5}

	for _, expectedActiveConns := range tests {
		c := fdhttp.NewClient()
		c.SetMaxIdleConns(expectedActiveConns)
		c.SetMaxIdleConnsPerHost(expectedActiveConns)

		var activeConns int32
		c.Use(middlewareConnCount(&activeConns))

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Millisecond)
		}))
		defer ts.Close()

		httpGetParallel(t, c, ts.URL, 10)
		assert.EqualValues(t, expectedActiveConns, atomic.LoadInt32(&activeConns))
	}
}
