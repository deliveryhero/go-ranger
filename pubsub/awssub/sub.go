package awssub

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/foodora/go-ranger/pubsub"
)

type (
	// subscriber is an SQS client that allows a user to
	// consume messages via the pubsub.Subscriber interface.
	subscriber struct {
		sqs sqsiface.SQSAPI

		cfg      SQSConfig
		queueURL *string

		stopped uint32

		stop   chan struct{}
		sqsErr error

		Logger pubsub.Logger

		// onErrorFunc is a func is being called when an error occurs
		onErrorFunc func(error)
	}

	// SQSMessage is the SQS implementation of `SubscriberMessage`.
	subscriberMessage struct {
		sub     *subscriber
		message *sqs.Message
	}
)

var (
	// defaultSQSMaxMessages is default the number of bulk messages
	// the subscriber will attempt to fetch on each
	// receive.
	defaultSQSMaxMessages int64 = 10
	// defaultSQSTimeoutSeconds is the default number of seconds the
	// SQS client will wait before timing out.
	defaultSQSTimeoutSeconds int64 = 2
	// defaultSQSSleepInterval is the default time.Duration the
	// subscriber will wait if it sees no messages
	// on the queue.
	defaultSQSSleepInterval = 2 * time.Second
)

var sqsClientFactoryFunc = createSqsClient

func defaultSQSConfig(cfg *SQSConfig) {
	if cfg.MaxMessages == 0 {
		cfg.MaxMessages = defaultSQSMaxMessages
	}

	if cfg.TimeoutSeconds == nil {
		cfg.TimeoutSeconds = &defaultSQSTimeoutSeconds
	}

	if cfg.SleepInterval == 0 {
		cfg.SleepInterval = defaultSQSSleepInterval
	}
}

// NewSubscriber will initiate a new Decrypter for the subscriber
// It will also fetch the SQS Queue Url
// and set up the SQS client.
func NewSubscriber(cfg SQSConfig) (pubsub.Subscriber, error) {
	var err error

	s := &subscriber{
		cfg:     cfg,
		stopped: 1,
		Logger:  pubsub.DefaultLogger,
	}

	if (len(cfg.QueueName) == 0) && (len(cfg.QueueURL) == 0) {
		return s, errors.New("sqs queue name or url is required")
	}

	sqsClient, err := sqsClientFactoryFunc(&cfg)
	if err != nil {
		return s, err
	}
	s.sqs = sqsClient

	if len(cfg.QueueURL) == 0 {
		var urlResp *sqs.GetQueueUrlOutput
		urlResp, err = s.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName:              &cfg.QueueName,
			QueueOwnerAWSAccountId: &cfg.QueueOwnerAccountID,
		})

		if err != nil {
			return s, err
		}

		s.queueURL = urlResp.QueueUrl
	} else {
		s.queueURL = &cfg.QueueURL
	}

	return s, nil
}

func createSqsClient(cfg *SQSConfig) (sqsiface.SQSAPI, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return sqs.New(sess, &cfg.Config), nil
}

// Message will decode message bodies and simply return string message.
func (m *subscriberMessage) String() string {
	msgBody := aws.StringValue(m.message.Body)
	return msgBody
}

// returns original sqs message id
func (m *subscriberMessage) GetMessageId() string {
	if m.message.MessageId == nil {
		return ""
	}
	return *m.message.MessageId
}

// ExtendDoneDeadline changes the visibility timeout of the underlying SQS
// message. It will set the visibility timeout of the message to the given
// duration.
func (m *subscriberMessage) ExtendDoneDeadline(d time.Duration) error {
	_, err := m.sub.sqs.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		QueueUrl:          m.sub.queueURL,
		ReceiptHandle:     m.message.ReceiptHandle,
		VisibilityTimeout: aws.Int64(int64(d.Seconds())),
	})
	return err
}

// Done removes a message from a queue
func (m *subscriberMessage) Done() error {
	deleteReq := &sqs.DeleteMessageInput{
		QueueUrl:      m.sub.queueURL,
		ReceiptHandle: m.message.ReceiptHandle,
	}
	_, err := m.sub.sqs.DeleteMessage(deleteReq)
	return err
}

// Returns the number of times a message has been received from the queue but not deleted.
func (m *subscriberMessage) GetReceiveCount() (int, error) {
	val, ok := m.message.Attributes[sqs.MessageSystemAttributeNameApproximateReceiveCount]
	if !ok || val == nil {
		return 0, errors.New("receive count is undefined")
	}
	n, err := strconv.Atoi(*val)
	if err != nil {
		return 0, fmt.Errorf("could not parse a string value '%s' to int", *val)
	}
	return n, nil
}

// Start will start consuming messages on the SQS queue
// and emit any messages to the returned channel.
// If it encounters any issues, it will populate the Err() error
// and close the returned channel.
func (s *subscriber) Start() <-chan pubsub.Message {
	if !s.isStopped() {
		err := errors.New("subscriber already is running")
		s.sqsErr = err
		s.onErrorFunc(err)
		return nil
	}
	atomic.SwapUint32(&s.stopped, uint32(0))
	s.stop = make(chan struct{})
	output := make(chan pubsub.Message)

	go func() {
		defer close(output)
		var (
			resp *sqs.ReceiveMessageOutput
			err  error
		)
		for {
			select {
			case <-s.stop:
				return
			default:
			}
			// get messages
			nameApproximateReceiveCount := sqs.MessageSystemAttributeNameApproximateReceiveCount
			resp, err = s.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: aws.Int64(s.cfg.MaxMessages),
				QueueUrl:            s.queueURL,
				WaitTimeSeconds:     s.cfg.TimeoutSeconds,
				AttributeNames:      []*string{&nameApproximateReceiveCount},
			})
			if err != nil {
				// we've encountered a major error
				s.Logger.Printf("Error occurred %s", err.Error())
				s.sqsErr = err
				if s.onErrorFunc != nil {
					s.onErrorFunc(err)
				}
				time.Sleep(s.cfg.SleepInterval)
				continue
			}

			// if we didn't get any messages, lets chill out for a sec
			if len(resp.Messages) == 0 {
				time.Sleep(s.cfg.SleepInterval)
				continue
			}

			// for each message, pass to output
			for _, msg := range resp.Messages {
				select {
				case <-s.stop:
					return
				case output <- &subscriberMessage{
					sub:     s,
					message: msg,
				}:
					continue
				}
			}
		}
	}()
	return output
}

// OnErrorFunc sets subscriber's onErrorFunc field
func (s *subscriber) SetOnErrorFunc(fn func(error)) {
	s.onErrorFunc = fn
}

func (s *subscriber) isStopped() bool {
	return atomic.LoadUint32(&s.stopped) == 1
}

// Stop will block until the consumer has stopped consuming
// messages.
func (s *subscriber) Stop() error {
	if s.isStopped() {
		return errors.New("sqs subscriber is not running")
	}
	atomic.SwapUint32(&s.stopped, uint32(1))
	// stop subscriber
	s.stop <- struct{}{}
	close(s.stop)
	return nil
}

// Err will contain any errors that occurred during
// consumption. This method should be checked after
// a user encounters a closed channel.
func (s *subscriber) Err() error {
	return s.sqsErr
}
