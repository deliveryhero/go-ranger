package fdapm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	newrelic "github.com/newrelic/go-agent"
)

type NewRelicTransaction struct {
	http.ResponseWriter

	EndInvoked             bool
	IgnoreInvoked          bool
	SetNameInvoked         bool
	NoticeErrorInvoked     bool
	AddAttributeInvoked    bool
	StartSegmentNowInvoked bool
}

func NewRelicTransactionMock(t *testing.T) *NewRelicTransaction {
	return &NewRelicTransaction{
		httptest.NewRecorder(),
	}
}

func (t *NewRelicTransaction) End() error {
	t.EndInvoked = true
	return nil
}

func (t *NewRelicTransaction) Ignore() error {
	t.IgnoreInvoked = true
	return nil
}

func (t *NewRelicTransaction) SetName(name string) error {
	t.SetNameInvoked = true
	return nil
}

func (t *NewRelicTransaction) NoticeError(err error) error {
	t.NoticeErrorInvoked = true
	return nil
}

func (t *NewRelicTransaction) AddAttribute(key string, value interface{}) error {
	t.AddAttributeInvoked = true
	return nil
}

func (t *NewRelicTransaction) StartSegmentNow() newrelic.SegmentStartTime {
	t.StartSegmentNowInvoked = true
	return newrelic.SegmentStartTime{}
}
