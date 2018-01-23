package ranger_metrics

import (
	"net/http"

	"github.com/foodora/go-ranger/ranger_logger"
	newrelic "github.com/newrelic/go-agent"
)

type NewRelic struct {
	Application newrelic.Application
}

func NewNewRelic(appName string, license string, logger ranger_logger.LoggerInterface) *NewRelic {
	app, err := newrelic.NewApplication(
		newrelic.NewConfig(
			appName,
			license,
		),
	)

	if err != nil {
		logger.Panic("NewRelic error", ranger_logger.LoggerData{"error": err.Error()})
	}

	return &NewRelic{
		Application: app,
	}
}

func (newRelic *NewRelic) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		txn := newRelic.Application.StartTransaction(r.URL.Path, w, r)
		defer txn.End()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// StartTransaction start a transaction manually
// call the returned function to end the transaction
func (newRelic *NewRelic) StartTransaction(w http.ResponseWriter, r *http.Request) func() {
	txn := newRelic.Application.StartTransaction(r.URL.Path, w, r)

	return func() {
		txn.End()
	}
}
