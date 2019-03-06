package awssub

import (
	"github.com/aws/aws-sdk-go/aws"
	"time"
)

// SQSConfig holds the info required to work with Amazon SQS
type SQSConfig struct {
	aws.Config
	QueueName           string
	QueueOwnerAccountID string
	// QueueURL can be used instead of QueueName and QueueOwnerAccountID.
	// If provided, the client will skip the "GetQueueUrl" call to AWS.
	QueueURL string
	// MaxMessages will override the DefaultSQSMaxMessages.
	MaxMessages int64
	// TimeoutSeconds will override the DefaultSQSTimeoutSeconds.
	TimeoutSeconds *int64
	// SleepInterval will override the DefaultSQSSleepInterval.
	SleepInterval time.Duration
	// DeleteBufferSize will override the DefaultSQSDeleteBufferSize.
	DeleteBufferSize *int
}

// NewSQSConfig return a SQSConfig instance to work with
func NewSQSConfig(config aws.Config) SQSConfig {
	return SQSConfig{
		Config: config,
	}
}
