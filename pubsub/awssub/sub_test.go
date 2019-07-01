package awssub

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/foodora/go-ranger/pubsub"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSubscriberForPositiveCases(t *testing.T) {
	test1 := "This is test 1"
	test2 := "This is test 2"
	test3 := "This is test 3"
	test4 := "This is test 4"
	test5 := "This is test 5"
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          &test1,
					ReceiptHandle: &test1,
				},
				{
					Body:          &test2,
					ReceiptHandle: &test2,
				},
			},
			{
				{
					Body:          &test3,
					ReceiptHandle: &test3,
				},
				{
					Body:          &test4,
					ReceiptHandle: &test4,
				},
			},
			{
				{
					Body:          &test5,
					ReceiptHandle: &test5,
				},
			},
		},
	}

	cfg := SQSConfig{
		QueueURL: "http://test_queue",
	}
	defaultSQSConfig(&cfg)
	cfg.DeleteBufferSize = aws.Int(2)
	sub, err := createSubscriber(cfg, sqstest)
	if err != nil {
		t.Error(err)
		return
	}

	queue := sub.Start()
	msq1 := <-queue
	verifyReceivedMsg(t, msq1, test1)
	msq1.Done()
	assert.True(t, len(sqstest.Deleted) == 0, "Message unexpectedly was removed from the delete buffer")

	msq2 := <-queue
	verifyReceivedMsg(t, msq2, test2)
	msq2.Done()
	assert.True(t, len(sqstest.Deleted) == 0, "Message unexpectedly was removed from the delete buffer")

	msq3 := <-queue
	verifyReceivedMsg(t, msq3, test3)
	msq3.Done()

	assert.True(t, len(sqstest.Deleted) == 3, "Messages were not removed from the delete buffer")
	verifyRemovedMsg(t, sqstest, msq1, 0)
	verifyRemovedMsg(t, sqstest, msq2, 1)
	verifyRemovedMsg(t, sqstest, msq3, 2)

	msq4 := <-queue
	verifyReceivedMsg(t, msq4, test4)
	msq4.Done()

	sub.Stop()

	assert.True(t, len(sqstest.Deleted) == 4, "The delete buffer was not flushed after the subscriber is stopped")
	verifyRemovedMsg(t, sqstest, msq4, 3)

	queue = sub.Start()
	msq5 := <-queue
	verifyReceivedMsg(t, msq5, test5)
	msq5.Done()

	sub.Stop()

	assert.True(t, len(sqstest.Deleted) == 5, "The delete buffer was not flushed after the subscriber is stopped")
	verifyRemovedMsg(t, sqstest, msq5, 4)

}

func TestSQSDoneAfterStop(t *testing.T) {
	test := "it stopped??"
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          &test,
					ReceiptHandle: &test,
				},
			},
		},
	}
	cfg := SQSConfig{
		QueueURL: "http://test_queue",
	}
	defaultSQSConfig(&cfg)
	sub, err := createSubscriber(cfg, sqstest)
	if err != nil {
		t.Error(err)
		return
	}

	queue := sub.Start()
	// verify we can receive a message, stop and still mark the message as 'done'
	gotRaw := <-queue
	sub.Stop()
	gotRaw.Done()
	// do all the other normal verifications
	if len(sqstest.Deleted) != 1 {
		t.Errorf("subscriber expected %d deleted message, got: %d", 1, len(sqstest.Deleted))
	}

	if *sqstest.Deleted[0].ReceiptHandle != test {
		t.Errorf("subscriber expected receipt handle of \"%s\" , got:+ \"%s\"",
			test,
			*sqstest.Deleted[0].ReceiptHandle)
	}
}

func TestExtendDoneTimeout(t *testing.T) {
	test := "some test"
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          &test,
					ReceiptHandle: &test,
				},
			},
		},
	}
	cfg := SQSConfig{
		QueueURL: "http://test_queue",
	}
	defaultSQSConfig(&cfg)
	sub, err := createSubscriber(cfg, sqstest)
	if err != nil {
		t.Error(err)
		return
	}

	queue := sub.Start()
	defer sub.Stop()
	gotRaw := <-queue
	gotRaw.ExtendDoneDeadline(time.Hour)
	if len(sqstest.Extended) != 1 {
		t.Errorf("subscriber expected %d extended message, got %d", 1, len(sqstest.Extended))
	}

	if *sqstest.Extended[0].ReceiptHandle != test {
		t.Errorf("subscriber expected receipt handle of %q , got:+ %q", test, *sqstest.Extended[0].ReceiptHandle)
	}
}

func createSubscriber(cfg SQSConfig, sqstest sqsiface.SQSAPI) (pubsub.Subscriber, error) {
	sqsClientFactoryFunc = func(cfg *SQSConfig) (sqsiface.SQSAPI, error) {
		return sqstest, nil
	}

	return NewSubscriber(cfg)
}

func verifyReceivedMsg(t *testing.T, msg pubsub.Message, expBody string) {
	assert.True(
		t,
		msg != nil,
		"subscriber did not receive a message '%s'", expBody)

	assert.Equal(
		t,
		msg.String(),
		expBody,
		"subscriber expected:\n%#v\ngot:\n%#v", expBody, msg.String())
}

func verifyRemovedMsg(t *testing.T, testsqs *TestSQSAPI, msg pubsub.Message, index int) {
	assert.True(
		t,
		len(testsqs.Deleted) > index,
		"Message '%s' was not removed", msg.String())

	assert.Equal(
		t,
		*testsqs.Deleted[index].ReceiptHandle,
		msg.String(),
		"Message '%s' was not removed", msg.String())
}

type TestSQSAPI struct {
	Offset   int
	Messages [][]*sqs.Message
	Deleted  []*sqs.DeleteMessageBatchRequestEntry
	Extended []*sqs.ChangeMessageVisibilityInput
	Err      error
}

var _ sqsiface.SQSAPI = &TestSQSAPI{}
var errNotImpl = errors.New("Not implemented ")

func (s *TestSQSAPI) ReceiveMessage(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if s.Offset >= len(s.Messages) {
		return &sqs.ReceiveMessageOutput{}, s.Err
	}
	out := s.Messages[s.Offset]
	s.Offset++
	return &sqs.ReceiveMessageOutput{Messages: out}, s.Err
}

func (s *TestSQSAPI) DeleteMessageBatch(i *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	s.Deleted = append(s.Deleted, i.Entries...)
	return nil, errNotImpl
}

func (s *TestSQSAPI) ChangeMessageVisibility(i *sqs.ChangeMessageVisibilityInput) (*sqs.ChangeMessageVisibilityOutput, error) {
	s.Extended = append(s.Extended, i)
	return nil, nil
}

///////////
// ALL METHODS BELOW HERE ARE EMPTY AND JUST SATISFYING THE SQSAPI interface
///////////

func (s *TestSQSAPI) DeleteMessage(d *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) DeleteMessageWithContext(aws.Context, *sqs.DeleteMessageInput, ...request.Option) (*sqs.DeleteMessageOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) DeleteMessageBatchRequest(i *sqs.DeleteMessageBatchInput) (*request.Request, *sqs.DeleteMessageBatchOutput) {
	return nil, nil
}
func (s *TestSQSAPI) DeleteMessageBatchWithContext(aws.Context, *sqs.DeleteMessageBatchInput, ...request.Option) (*sqs.DeleteMessageBatchOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) AddPermissionRequest(*sqs.AddPermissionInput) (*request.Request, *sqs.AddPermissionOutput) {
	return nil, nil
}
func (s *TestSQSAPI) AddPermission(*sqs.AddPermissionInput) (*sqs.AddPermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) AddPermissionWithContext(aws.Context, *sqs.AddPermissionInput, ...request.Option) (*sqs.AddPermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ChangeMessageVisibilityRequest(*sqs.ChangeMessageVisibilityInput) (*request.Request, *sqs.ChangeMessageVisibilityOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ChangeMessageVisibilityWithContext(aws.Context, *sqs.ChangeMessageVisibilityInput, ...request.Option) (*sqs.ChangeMessageVisibilityOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) ChangeMessageVisibilityBatchRequest(*sqs.ChangeMessageVisibilityBatchInput) (*request.Request, *sqs.ChangeMessageVisibilityBatchOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ChangeMessageVisibilityBatch(*sqs.ChangeMessageVisibilityBatchInput) (*sqs.ChangeMessageVisibilityBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ChangeMessageVisibilityBatchWithContext(aws.Context, *sqs.ChangeMessageVisibilityBatchInput, ...request.Option) (*sqs.ChangeMessageVisibilityBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) CreateQueueRequest(*sqs.CreateQueueInput) (*request.Request, *sqs.CreateQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) CreateQueue(*sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) CreateQueueWithContext(aws.Context, *sqs.CreateQueueInput, ...request.Option) (*sqs.CreateQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) DeleteMessageRequest(*sqs.DeleteMessageInput) (*request.Request, *sqs.DeleteMessageOutput) {
	return nil, nil
}

func (s *TestSQSAPI) DeleteQueueRequest(*sqs.DeleteQueueInput) (*request.Request, *sqs.DeleteQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) DeleteQueue(*sqs.DeleteQueueInput) (*sqs.DeleteQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) DeleteQueueWithContext(aws.Context, *sqs.DeleteQueueInput, ...request.Option) (*sqs.DeleteQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueAttributesRequest(*sqs.GetQueueAttributesInput) (*request.Request, *sqs.GetQueueAttributesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) GetQueueAttributes(*sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueAttributesWithContext(aws.Context, *sqs.GetQueueAttributesInput, ...request.Option) (*sqs.GetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueUrlRequest(*sqs.GetQueueUrlInput) (*request.Request, *sqs.GetQueueUrlOutput) {
	return nil, nil
}
func (s *TestSQSAPI) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueUrlWithContext(aws.Context, *sqs.GetQueueUrlInput, ...request.Option) (*sqs.GetQueueUrlOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListDeadLetterSourceQueuesRequest(*sqs.ListDeadLetterSourceQueuesInput) (*request.Request, *sqs.ListDeadLetterSourceQueuesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ListDeadLetterSourceQueues(*sqs.ListDeadLetterSourceQueuesInput) (*sqs.ListDeadLetterSourceQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListDeadLetterSourceQueuesWithContext(aws.Context, *sqs.ListDeadLetterSourceQueuesInput, ...request.Option) (*sqs.ListDeadLetterSourceQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueuesRequest(*sqs.ListQueuesInput) (*request.Request, *sqs.ListQueuesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ListQueues(*sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueuesWithContext(aws.Context, *sqs.ListQueuesInput, ...request.Option) (*sqs.ListQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) PurgeQueueRequest(*sqs.PurgeQueueInput) (*request.Request, *sqs.PurgeQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) PurgeQueue(*sqs.PurgeQueueInput) (*sqs.PurgeQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) PurgeQueueWithContext(aws.Context, *sqs.PurgeQueueInput, ...request.Option) (*sqs.PurgeQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ReceiveMessageRequest(*sqs.ReceiveMessageInput) (*request.Request, *sqs.ReceiveMessageOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) RemovePermissionRequest(*sqs.RemovePermissionInput) (*request.Request, *sqs.RemovePermissionOutput) {
	return nil, nil
}
func (s *TestSQSAPI) RemovePermission(*sqs.RemovePermissionInput) (*sqs.RemovePermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) RemovePermissionWithContext(aws.Context, *sqs.RemovePermissionInput, ...request.Option) (*sqs.RemovePermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageRequest(*sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
	return nil, nil
}
func (s *TestSQSAPI) SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageWithContext(aws.Context, *sqs.SendMessageInput, ...request.Option) (*sqs.SendMessageOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageBatchRequest(*sqs.SendMessageBatchInput) (*request.Request, *sqs.SendMessageBatchOutput) {
	return nil, nil
}
func (s *TestSQSAPI) SendMessageBatch(*sqs.SendMessageBatchInput) (*sqs.SendMessageBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageBatchWithContext(aws.Context, *sqs.SendMessageBatchInput, ...request.Option) (*sqs.SendMessageBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SetQueueAttributesRequest(*sqs.SetQueueAttributesInput) (*request.Request, *sqs.SetQueueAttributesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) SetQueueAttributes(*sqs.SetQueueAttributesInput) (*sqs.SetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SetQueueAttributesWithContext(aws.Context, *sqs.SetQueueAttributesInput, ...request.Option) (*sqs.SetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueueTags(input *sqs.ListQueueTagsInput) (*sqs.ListQueueTagsOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueueTagsRequest(input *sqs.ListQueueTagsInput) (req *request.Request, output *sqs.ListQueueTagsOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ListQueueTagsWithContext(ctx aws.Context, input *sqs.ListQueueTagsInput, opts ...request.Option) (*sqs.ListQueueTagsOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) TagQueue(input *sqs.TagQueueInput) (*sqs.TagQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) TagQueueRequest(input *sqs.TagQueueInput) (req *request.Request, output *sqs.TagQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) TagQueueWithContext(ctx aws.Context, input *sqs.TagQueueInput, opts ...request.Option) (*sqs.TagQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) UntagQueue(input *sqs.UntagQueueInput) (*sqs.UntagQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) UntagQueueRequest(input *sqs.UntagQueueInput) (req *request.Request, output *sqs.UntagQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) UntagQueueWithContext(ctx aws.Context, input *sqs.UntagQueueInput, opts ...request.Option) (*sqs.UntagQueueOutput, error) {
	return nil, errNotImpl
}
