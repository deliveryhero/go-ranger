package fdhttp

import (
	"io"
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
	http.Client
}

// DefaultClientTimeout will be used when create a new Client
var DefaultClientTimeout = 10 * time.Second

// NewClient return a new instace of fdhttp.Client
func NewClient() *ClientImpl {
	return &ClientImpl{
		Client: http.Client{
			Timeout:   DefaultClientTimeout,
			Transport: http.DefaultTransport,
		},
	}
}

// Use a middleware to wrap all http calls
func (c *ClientImpl) Use(middlewares ...fdmiddleware.ClientMiddleware) {
	for _, m := range middlewares {
		c.Client.Transport = m.Wrap(c.Client.Transport)
	}
}
