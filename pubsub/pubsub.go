package pubsub

import (
	"context"
	"time"
)

// Publisher ...
type Publisher interface {
	// Publish will publish a message with context.
	Publish(context.Context, string, string) error

	// Publish will publish a message with context.
	PublishToTopic(ctx context.Context, key string, m string, topic string) error
}

// Subscriber ...
type Subscriber interface {
	// Start will return a channel of raw messages.
	Start() <-chan Message
	// Err will contain any errors returned from the consumer connection.
	Err() error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
	// SetOnErrorFunc is a setter for a func is being called when an error occurs
	SetOnErrorFunc(fn func(error))
}

// Message ...
type Message interface {
	String() string
	ExtendDoneDeadline(time.Duration) error
	Done() error
	GetReceiveCount() (int, error)
	GetMessageId() string
}
