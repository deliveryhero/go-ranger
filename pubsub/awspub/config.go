package awspub

import "github.com/aws/aws-sdk-go/aws"

// SNSConfig holds the info required to work with Amazon SNS.
type SNSConfig struct {
	aws.Config

	Topic string
}

// NewSNSConfig return a SNSConfig instance to work with
func NewSNSConfig(config aws.Config, topic string) SNSConfig {
	return SNSConfig{
		config,
		topic,
	}
}
