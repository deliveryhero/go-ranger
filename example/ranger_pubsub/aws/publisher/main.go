package main

import (
	"context"
	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/foodora/go-ranger/ranger_pubsub/aws"
	"log"
)

// An example application to create a publisher and publish a message to SNS
func main() {
	// create configuration for publisher
	config := aws.SNSConfig{
		Config: *aws2.NewConfig().WithRegion("eu-central-1"),
		Topic:  "<topic-arn>",
	}

	// initialize a publisher instance
	publisher, err := aws.NewPublisher(config)
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
