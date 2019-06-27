package awssub

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/foodora/go-ranger/pubsub"
	"sync/atomic"
	"time"
)

type (
	// subscriber is an SQS client that allows a user to
	// consume messages via the pubsub.Subscriber interface.
	subscriber struct {
		sqs sqsiface.SQSAPI

		cfg      SQSConfig
		queueURL *string

		toDelete chan *deleteRequest
		// inFlight and stopped are signals to manage delete requests
		// at shutdown.
		inFlight uint64
		stopped  uint32

		stop   chan chan error
		sqsErr error

		Logger pubsub.Logger
	}

	// SQSMessage is the SQS implementation of `SubscriberMessage`.
	subscriberMessage struct {
		sub     *subscriber
		message *sqs.Message
	}

	deleteRequest struct {
		entry   *sqs.DeleteMessageBatchRequestEntry
		receipt chan error
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

	// defaultSQSDeleteBufferSize is the default limit of messages
	// allowed in the delete buffer before
	// executing a 'delete batch' request.
	defaultSQSDeleteBufferSize = 0

	defaultSQSConsumeBase64 = true
)

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

	if cfg.DeleteBufferSize == nil {
		cfg.DeleteBufferSize = &defaultSQSDeleteBufferSize
	}
}

// incrementInflight will increment the add in flight count.
func (s *subscriber) incrementInFlight() {
	atomic.AddUint64(&s.inFlight, 1)
}

// removeInfFlight will decrement the in flight count.
func (s *subscriber) decrementInFlight() {
	atomic.AddUint64(&s.inFlight, ^uint64(0))
}

// inFlightCount returns the number of in-flight requests currently
// running on this server.
func (s *subscriber) inFlightCount() uint64 {
	return atomic.LoadUint64(&s.inFlight)
}

// NewSubscriber will initiate a new Decrypter for the subscriber
// It will also fetch the SQS Queue Url
// and set up the SQS client.
func NewSubscriber(cfg SQSConfig) (pubsub.Subscriber, error) {
	var err error

	s := &subscriber{
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
		Logger:   pubsub.DefaultLogger,
	}

	if (len(cfg.QueueName) == 0) && (len(cfg.QueueURL) == 0) {
		return s, errors.New("sqs queue name or url is required")
	}

	sess, err := session.NewSession()
	if err != nil {
		return s, err
	}

	s.sqs = sqs.New(sess, &aws.Config{
		Region:   cfg.Region,
		Endpoint: cfg.Endpoint,
	})

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

// Done will queue up a message to be deleted. By default,
// the `SQSDeleteBufferSize` will be 0, so this will block until the
// message has been deleted.
func (m *subscriberMessage) Done() error {
	defer m.sub.decrementInFlight()
	receipt := make(chan error)
	m.sub.toDelete <- &deleteRequest{
		entry: &sqs.DeleteMessageBatchRequestEntry{
			Id:            m.message.MessageId,
			ReceiptHandle: m.message.ReceiptHandle,
		},
		receipt: receipt,
	}
	return <-receipt
}

// Start will start consuming messages on the SQS queue
// and emit any messages to the returned channel.
// If it encounters any issues, it will populate the Err() error
// and close the returned channel.
func (s *subscriber) Start() <-chan pubsub.Message {
	output := make(chan pubsub.Message)
	go s.handleDeletes()
	go func() {
		defer close(output)
		var (
			resp *sqs.ReceiveMessageOutput
			err  error
		)
		for {
			select {
			case exit := <-s.stop:
				exit <- nil
				return
			default:
			}
			s.Logger.Printf("receiving messages")
			// get messages
			resp, err = s.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: aws.Int64(s.cfg.MaxMessages),
				QueueUrl:            s.queueURL,
				WaitTimeSeconds:     s.cfg.TimeoutSeconds,
			})
			if err != nil {
				// we've encountered a major error
				s.Logger.Printf("Error occurred %s", err.Error())
				s.sqsErr = err
				time.Sleep(s.cfg.SleepInterval)
				continue
			}

			// if we didn't get any messages, lets chill out for a sec
			if len(resp.Messages) == 0 {
				s.Logger.Printf("no messages found. sleeping for %s", s.cfg.SleepInterval)
				time.Sleep(s.cfg.SleepInterval)
				continue
			}

			s.Logger.Printf("found %d messages", len(resp.Messages))
			// for each message, pass to output
			for _, msg := range resp.Messages {
				output <- &subscriberMessage{
					sub:     s,
					message: msg,
				}
				s.incrementInFlight()
			}
		}
	}()
	return output
}

func (s *subscriber) handleDeletes() {
	batchInput := &sqs.DeleteMessageBatchInput{
		QueueUrl: s.queueURL,
	}
	var (
		err           error
		entriesBuffer []*sqs.DeleteMessageBatchRequestEntry
		delRequest    *deleteRequest
	)
	for delRequest = range s.toDelete {
		entriesBuffer = append(entriesBuffer, delRequest.entry)
		// if the subscriber is stopped and this is the last request,
		// flush quit!
		if s.isStopped() && s.inFlightCount() == 1 {
			break
		}
		// if buffer is full, send the request
		if len(entriesBuffer) > *s.cfg.DeleteBufferSize {
			batchInput.Entries = entriesBuffer
			_, err = s.sqs.DeleteMessageBatch(batchInput)
			// clear buffer
			entriesBuffer = []*sqs.DeleteMessageBatchRequestEntry{}
		}

		delRequest.receipt <- err
	}
	// clear any remainders before shutdown
	if len(entriesBuffer) > 0 {
		batchInput.Entries = entriesBuffer
		_, err = s.sqs.DeleteMessageBatch(batchInput)
		delRequest.receipt <- err
	}
}

func (s *subscriber) isStopped() bool {
	return atomic.LoadUint32(&s.stopped) == 1
}

// Stop will block until the consumer has stopped consuming
// messages.
func (s *subscriber) Stop() error {
	if s.isStopped() {
		return errors.New("sqs subscriber is already stopped")
	}
	exit := make(chan error)
	s.stop <- exit
	atomic.SwapUint32(&s.stopped, uint32(1))
	return <-exit
}

// Err will contain any errors that occurred during
// consumption. This method should be checked after
// a user encounters a closed channel.
func (s *subscriber) Err() error {
	return s.sqsErr
}
