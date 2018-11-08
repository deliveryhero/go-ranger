package fdhttp

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}

// Client is a wrap to http.Client where you can add different configurations, like
// fdhttp.WithFallback(), fdhttp.WithBackoff(), etc.
type ClientImpl struct {
	*http.Client
	// Control when abort ticker to close idle connections.
	maxLifetimeDone chan struct{}
}

// DefaultClientTimeout will be used when create a new Client
var DefaultClientTimeout = 10 * time.Second

// NewClient return a new instace of fdhttp.Client
func NewClient() *ClientImpl {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &ClientImpl{
		Client: &http.Client{
			Timeout:   DefaultClientTimeout,
			Transport: tr,
		},
	}
}

func (c *ClientImpl) httpTransport() *http.Transport {
	tr, _ := c.Transport.(*http.Transport)
	return tr
}

// Use a middleware to wrap all http calls
func (c *ClientImpl) Use(middlewares ...fdmiddleware.ClientMiddleware) {
	for _, m := range middlewares {
		c.Client.Transport = m.Wrap(c.Client.Transport)
	}
}

// StdClient return the http.Client from standard library will all
// configuration that you have changed. Use it if you need to communicate
// with different libraries.
func (c *ClientImpl) StdClient() *http.Client {
	return c.Client
}

// SetMaxIdleConns controls the maximum number of idle (keep-alive)
// connections across all hosts. Zero means no limit.
func (c *ClientImpl) SetMaxIdleConns(n int) {
	tr := c.httpTransport()
	if tr != nil {
		tr.MaxIdleConns = n
	}
}

// SetMaxIdleConnsPerHost if non-zero, controls the maximum idle
// (keep-alive) connections to keep per-host. If zero,
// http.DefaultMaxIdleConnsPerHost is used.
func (c *ClientImpl) SetMaxIdleConnsPerHost(n int) {
	tr := c.httpTransport()
	if tr != nil {
		tr.MaxIdleConnsPerHost = n
	}
}

// SetIdleConnTimeout sets the maximum amount of time an idle
// (keep-alive) connection will remain idle before closing
// itself. Zero means no limit.
func (c *ClientImpl) SetIdleConnTimeout(d time.Duration) {
	tr := c.httpTransport()
	if tr != nil {
		tr.IdleConnTimeout = d
	}
}

// SetIdleConnMaxLifetime sets the maximum amount of time a connection may be reused.
// If d <= 0, connections are reused forever.
func (c *ClientImpl) SetIdleConnMaxLifetime(d time.Duration) {
	if c.maxLifetimeDone != nil {
		c.maxLifetimeDone <- struct{}{}
		c.maxLifetimeDone = nil
	}

	tr := c.httpTransport()
	if tr == nil {
		return
	}
	if d <= 0 {
		return
	}

	c.maxLifetimeDone = make(chan struct{})
	t := time.NewTicker(d)

	go func() {
		defer t.Stop()
		for {
			select {
			case <-t.C:
				tr.CloseIdleConnections()
			case <-c.maxLifetimeDone:
				return
			}
		}
	}()
}
