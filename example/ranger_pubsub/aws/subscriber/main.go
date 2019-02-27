package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/foodora/go-ranger/pubsub/awssub"
	"log"
)

// An example application to create a subscriber and listening on to a message from subscriber
func main() {
	awsconfig := *aws.NewConfig().WithRegion("eu-central-1")
	// create configuration for publisher
	config := awssub.NewSQSConfig(awsconfig)
	config.QueueName = "<name-of-the-sqs-queue>"
	config.QueueURL = "<sqs-queue-url>" //optional
	config.MaxMessages = 10
	config.TimeoutSeconds = aws.Int64(10)

	// initialize a subscriber instance
	subscriber, err := awssub.NewSubscriber(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// start pooling messages, this will return channel of messages
	messageQueue := subscriber.Start()

	// try reading message from the queue
	rawMessage := <-messageQueue

	//process the message as per the business need
	message := rawMessage.String()
	fmt.Println(message)

	// queue up the message from
	rawMessage.Done()

	// stop polling messages, shutdown the subscriber connection
	err = subscriber.Stop()
	if err != nil {
		log.Fatal()
	}
}
