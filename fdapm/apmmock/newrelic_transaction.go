package apmmock

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	newrelic "github.com/newrelic/go-agent"
)

type NewRelicTransaction struct {
	mu sync.Mutex

	http.ResponseWriter
	EndInvoked                           bool
	IgnoreInvoked                        bool
	SetNameInvoked                       bool
	NoticeErrorInvoked                   bool
	AddAttributeInvoked                  bool
	StartSegmentNowInvoked               bool
	AcceptDistributedTracePayloadInvoked bool
	CreateDistributedTracePayloadInvoked bool
}

func NewNRTransaction(t *testing.T) *NewRelicTransaction {
	return &NewRelicTransaction{
		ResponseWriter: httptest.NewRecorder(),
	}
}

func (t *NewRelicTransaction) End() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.EndInvoked = true
	return nil
}

func (t *NewRelicTransaction) Ignore() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.IgnoreInvoked = true
	return nil
}

func (t *NewRelicTransaction) SetName(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.SetNameInvoked = true
	return nil
}

func (t *NewRelicTransaction) NoticeError(err error) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.NoticeErrorInvoked = true
	return nil
}

func (t *NewRelicTransaction) AddAttribute(key string, value interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.AddAttributeInvoked = true
	return nil
}

func (t *NewRelicTransaction) StartSegmentNow() newrelic.SegmentStartTime {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.StartSegmentNowInvoked = true
	return newrelic.SegmentStartTime{}
}

func (t *NewRelicTransaction) AcceptDistributedTracePayload(newrelic.TransportType, interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.AcceptDistributedTracePayloadInvoked = true
	return nil
}

func (t *NewRelicTransaction) CreateDistributedTracePayload() newrelic.DistributedTracePayload {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.CreateDistributedTracePayloadInvoked = true
	return mockPayload{}
}

type mockPayload struct{}

func (p mockPayload) Text() string     { return "" }
func (p mockPayload) HTTPSafe() string { return "" }
