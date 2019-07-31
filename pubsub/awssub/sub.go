package awssub

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/foodora/go-ranger/pubsub"
	"strconv"
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
		flush    chan chan error
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

	sqsClient := sqs.New(sess, &aws.Config{
		Region:   cfg.Region,
		Endpoint: cfg.Endpoint,
	})

	return sqsClient, nil
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
	batchInput := &sqs.DeleteMessageBatchRequestEntry{
		Id:            m.message.MessageId,
		ReceiptHandle: m.message.ReceiptHandle,
	}
	if m.sub.isStopped() {
		return m.sub.deleteMessageBatch(batchInput)
	}
	receipt := make(chan error)
	m.sub.toDelete <- &deleteRequest{
		entry:   batchInput,
		receipt: receipt,
	}
	return <-receipt
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
		s.sqsErr = errors.New("subscriber already is running")
		return nil
	}
	atomic.SwapUint32(&s.stopped, uint32(0))
	s.stop = make(chan chan error, 1)
	s.flush = make(chan chan error, 1)
	s.toDelete = make(chan *deleteRequest)

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
				select {
				case exit := <-s.stop:
					exit <- nil
					return
				case output <- &subscriberMessage{
					sub:     s,
					message: msg,
				}:
					s.incrementInFlight()
				}
			}
		}
	}()
	return output
}

func (s *subscriber) handleDeletes() {
	var entriesBuffer []*sqs.DeleteMessageBatchRequestEntry
	for {
		var delRequest *deleteRequest
		var err error
		select {
		case flush := <-s.flush:
			if len(entriesBuffer) > 0 {
				err = s.deleteMessageBatch(entriesBuffer...)
			}
			flush <- err
			return
		case delRequest = <-s.toDelete:
		}
		entriesBuffer = append(entriesBuffer, delRequest.entry)
		// if buffer is full, send the request
		if len(entriesBuffer) > *s.cfg.DeleteBufferSize {
			err = s.deleteMessageBatch(entriesBuffer...)
			entriesBuffer = []*sqs.DeleteMessageBatchRequestEntry{}
		}
		delRequest.receipt <- err
	}
}

// deleteMessageBatch is helper function to remove messages from sqs in batches
func (s *subscriber) deleteMessageBatch(batchReq ...*sqs.DeleteMessageBatchRequestEntry) error {
	batchInput := &sqs.DeleteMessageBatchInput{
		QueueUrl: s.queueURL,
		Entries:  batchReq,
	}
	_, err := s.sqs.DeleteMessageBatch(batchInput)

	return err
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
	defer func() {
		close(s.stop)
		close(s.toDelete)
		close(s.flush)
	}()
	exit := make(chan error)
	defer close(exit)
	// stop subscriber
	s.stop <- exit
	atomic.SwapUint32(&s.stopped, uint32(1))
	err := <-exit
	if err != nil {
		return err
	}
	//flush deleted msg buffer
	flush := make(chan error)
	defer close(flush)
	s.flush <- flush

	return <-flush
}

// Err will contain any errors that occurred during
// consumption. This method should be checked after
// a user encounters a closed channel.
func (s *subscriber) Err() error {
	return s.sqsErr
}
