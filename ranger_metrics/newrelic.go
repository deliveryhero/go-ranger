package ranger_metrics

import (
	"errors"
	"net/http"

	"github.com/foodora/go-ranger/ranger_http"
	"github.com/foodora/go-ranger/ranger_logger"
	newrelic "github.com/newrelic/go-agent"
)

type NewRelic struct {
	appName     string
	license     string
	logger      ranger_logger.LoggerInterface
	Application newrelic.Application
}

func NewRelicLabels(middleware ranger_http.MiddlewareInterface, labels map[string]string) {
	nr := middleware.(*NewRelic)
	// create new application because we cannot change config of a created application
	config := newrelic.NewConfig(nr.appName, nr.license)
	config.Labels = labels
	config.ErrorCollector.IgnoreStatusCodes = []int{
		http.StatusBadRequest, // 400
	}
	config.CrossApplicationTracer.Enabled = false
	config.DistributedTracer.Enabled = true

	app, err := newrelic.NewApplication(config)
	if err != nil {
		nr.logger.Panic("NewRelic cannot create app with labels", ranger_logger.LoggerData{"error": err.Error()})
	}

	nr.Application = app
}

func NewNewRelic(appName string, license string, logger ranger_logger.LoggerInterface) *NewRelic {
	config := newrelic.NewConfig(appName, license)

	return NewNewRelicWithConfig(config, logger)
}

func NewNewRelicWithConfig(config newrelic.Config, logger ranger_logger.LoggerInterface) *NewRelic {
	app, err := newrelic.NewApplication(config)
	if err != nil {
		logger.Panic("NewRelic error", ranger_logger.LoggerData{"error": err.Error()})
	}

	return &NewRelic{
		appName:     config.AppName,
		license:     config.License,
		logger:      logger,
		Application: app,
	}
}

func (newRelic *NewRelic) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		txn := newRelic.Application.StartTransaction(r.URL.Path, w, r)
		defer txn.End()

		ctx := newrelic.NewContext(r.Context(), txn)
		r = r.WithContext(ctx)

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
