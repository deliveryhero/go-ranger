package main

import (
	"fmt"
	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/foodora/go-ranger/ranger_pubsub/aws"
	"log"
)

// An example application to create a subscriber and listening on to a message from subscriber
func main() {
	// create configuration for publisher
	config := aws.SQSConfig{
		Config:         *aws2.NewConfig().WithRegion("eu-central-1"),
		QueueName:      "<name-of-the-sqs-queue>",
		QueueURL:       "<sqs-queue-url>", //optional
		MaxMessages:    aws2.Int64(10),
		TimeoutSeconds: aws2.Int64(10),
	}

	// initialize a subscriber instance
	subscriber, err := aws.NewSubscriber(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// start pooling messages, this will return channel of messages
	messageQueue := subscriber.Start()

	// try reading message from the queue
	rawMessage := <-messageQueue

	//process the message as per the business need
	message := string(rawMessage.Message())
	fmt.Println(message)

	// queue up the message from
	rawMessage.Done()

	// stop polling messages, shutdown the subscriber connection
	err = subscriber.Stop()
	if err != nil {
		log.Fatal()
	}
}
