package fdhttp

import (
	"net/http"
	"time"
)

// Client is a wrap to http.Client where you can add different configurations, like
// fdhttp.WithFallback(), fdhttp.WithBackoff(), etc.
type Client struct {
	http.Client

	middlewares []ClientMiddleware

	// Logger will be setted with DefaultLogger when NewClient is called
	// but you can overwrite later only in this instance.
	Logger Logger
}

// DefaultClientTimeout will be used when create a new Client
var DefaultClientTimeout = 10 * time.Second

// NewClient return a new instace of fdhttp.Client
func NewClient() *Client {
	return &Client{
		Client: http.Client{
			Timeout: DefaultClientTimeout,
		},
		Logger: defaultLogger,
	}
}

// Use a middleware to wrap all http calls
func (c *Client) Use(m ...ClientMiddleware) {
	c.middlewares = append(c.middlewares, m...)
}

// Do is the implementation to call all middlewares before efective do the call.
// It also make fdhttp.Client a valid Doer
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var reqDo Doer = &c.Client
	for k := range c.middlewares {
		reqDo = c.middlewares[len(c.middlewares)-1-k](reqDo)
	}

	return reqDo.Do(req)
}
