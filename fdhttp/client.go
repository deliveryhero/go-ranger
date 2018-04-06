package fdhttp

import (
	"net/http"
	"time"
)

type Client struct {
	http.Client

	// Logger will be setted with DefaultLogger when NewClient is called
	// but you can overwrite later only in this instance.
	Logger Logger
}

func NewClient() *Client {
	return &Client{
		Client: http.Client{},
		Logger: defaultLogger,
	}
}

func (c *Client) SetTimeout(d time.Duration) {
	c.Timeout = d
}
