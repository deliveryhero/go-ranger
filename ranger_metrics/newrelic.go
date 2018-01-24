package ranger_metrics

import (
	"errors"
	"net/http"

	"github.com/foodora/go-ranger/ranger_logger"
	newrelic "github.com/newrelic/go-agent"
)

type NewRelic struct {
	Application newrelic.Application
	logger      ranger_logger.LoggerInterface
}

func NewNewRelic(appName string, license string, logger ranger_logger.LoggerInterface) *NewRelic {
	config := newrelic.NewConfig(appName, license)

	app, err := newrelic.NewApplication(config)
	if err != nil {
		logger.Panic("NewRelic error", ranger_logger.LoggerData{"error": err.Error()})
	}

	return &NewRelic{
		Application: app,
		logger:      logger,
	}
}

func (newRelic *NewRelic) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		txn := newRelic.Application.StartTransaction(r.URL.Path, w, r)
		defer txn.End()

		next.ServeHTTP(txn, r)
	}

	return http.HandlerFunc(fn)
}

// StartTransaction start a transaction manually
// * Call the returned function to end the transaction
func (newRelic *NewRelic) StartTransaction(w http.ResponseWriter, r *http.Request) func() {
	txn := newRelic.Application.StartTransaction(r.URL.Path, w, r)

	return func() {
		txn.End()
	}
}

// NoticeError send content of err to newrelic using trasaction
// that middleware has started
func (newRelic *NewRelic) NoticeError(w http.ResponseWriter, err error) {
	if txn, ok := w.(newrelic.Transaction); ok {
		txnErr := txn.NoticeError(err)
		if txnErr != nil {
			newRelic.logger.Error("Cannot send error to NewRelic", ranger_logger.LoggerData{
				"error": err.Error(),
			})
		}

		return
	}

	newRelic.logger.Error("Cannot send error to NewRelic", ranger_logger.LoggerData{
		"error": errors.New("Transaction wasn't started by NewRelic Middleware"),
	})
}
