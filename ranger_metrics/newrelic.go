package ranger_metrics

import (
	"errors"
	"net/http"

	"github.com/foodora/go-ranger/ranger_logger"
	newrelic "github.com/newrelic/go-agent"
)

// NewRelic ...
type NewRelic struct {
	newrelic.Application
	ranger_logger.LoggerInterface
}

// NewNewRelic ...
func NewNewRelic(appName string, license string, logger ranger_logger.LoggerInterface) *NewRelic {
	config := newrelic.NewConfig(appName, license)

	app, err := newrelic.NewApplication(config)
	if err != nil {
		logger.Panic("NewRelic error", ranger_logger.LoggerData{"error": err.Error()})
	}

	return &NewRelic{
		Application:     app,
		LoggerInterface: logger,
	}
}

// Middleware ...
func (nr *NewRelic) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		txn := nr.Application.StartTransaction(r.URL.Path, w, r)
		defer txn.End()

		next.ServeHTTP(txn, r)
	}

	return http.HandlerFunc(fn)
}

// StartTransaction start a transaction manually
// * Call the returned function to end the transaction
func (nr *NewRelic) StartTransaction(w http.ResponseWriter, r *http.Request) func() {
	txn := nr.Application.StartTransaction(r.URL.Path, w, r)

	return func() {
		txn.End()
	}
}

// UseNewRoundTripper to have our outbound requests eligible for cross application tracing
func (nr *NewRelic) UseNewRoundTripper(txn newrelic.Transaction, client *http.Client) {
	client.Transport = newrelic.NewRoundTripper(txn, nil)
}

// NoticeError send content of err to newrelic using trasaction
// that middleware has started
func (nr *NewRelic) NoticeError(w http.ResponseWriter, err error) {
	if txn, ok := w.(newrelic.Transaction); ok {
		txnErr := txn.NoticeError(err)
		if txnErr != nil {
			nr.Error("Cannot send error to NewRelic", ranger_logger.LoggerData{
				"error": err.Error(),
			})
		}

		return
	}

	nr.Error("Cannot send error to NewRelic", ranger_logger.LoggerData{
		"error": errors.New("Transaction wasn't started by NewRelic Middleware"),
	})
}
