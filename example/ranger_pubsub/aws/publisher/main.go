package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/foodora/go-ranger/pubsub/awspub"

	"log"
)

// An example application to create a publisher and publish a message to SNS
func main() {
	awsconfig := *aws.NewConfig().WithRegion("eu-central-1")
	topic := "<topic-arn>"
	// create configuration for publisher
	config := awspub.NewSNSConfig(awsconfig, topic)

	// initialize a publisher instance
	publisher, err := awspub.NewPublisher(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// create subject and message
	subject := "this is a subject" //optional
	message := "this is a message"

	// publish a message
	err = publisher.Publish(context.Background(), subject, message)
	if err != nil {
		log.Fatal(err.Error())
	}
}
