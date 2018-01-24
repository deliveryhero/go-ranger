package ranger_metrics

import (
	"errors"
	"net/http"
	"testing"

	"github.com/foodora/go-ranger/ranger_logger"
)

func TestInitNewRelic(t *testing.T) {
	//t.Error("@todo TestInitNewRelic")
}

func ExampleStartTransactionManually() {
	var logger ranger_logger.LoggerInterface

	newRelic := NewNewRelic("appName", "license", logger)

	// Start and end manually
	closeTxn := newRelic.StartTransaction(nil, nil)
	// your code here
	closeTxn()

	// Start and end with defer
	closeTxn = newRelic.StartTransaction(nil, nil)
	defer closeTxn()
	// your code here

	// Start and end with defer in the same line
	defer newRelic.StartTransaction(nil, nil)()
	// your code here
}

func ExampleNoticeError() {
	var logger ranger_logger.LoggerInterface

	newRelic := NewNewRelic("appName", "license", logger)

	// Ensure that your handler is called by NewRelic.Middleware
	handler := newRelic.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newRelic.NoticeError(w, errors.New("Application failed!"))
	}))

	http.Handle("/foo", handler)
}
