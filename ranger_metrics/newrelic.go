package ranger_metrics

import (
	newrelic "github.com/newrelic/go-agent"
	"net/http"
	"github.com/fesposito/go-ranger/ranger_logger"
)

type NewRelic struct {
	Application newrelic.Application
}

func NewNewRelic(appName string, license string, logger ranger_logger.LoggerInterface) *NewRelic {
	app, err := newrelic.NewApplication(newrelic.NewConfig(
		appName,
		license),
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
