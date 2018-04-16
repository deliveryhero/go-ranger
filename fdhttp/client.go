package fdhttp

import (
	"net/http"
)

// Client is a wrap to http.Client where you can add different configurations, like
// fdhttp.WithFallback(), fdhttp.WithBackoff(), etc.
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
