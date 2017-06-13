package main

import (
	"encoding/json"
	"net/http"

	ranger_http "github.com/fesposito/go-ranger/ranger_http"
	ranger_logger "github.com/fesposito/go-ranger/ranger_logger"
	ranger_metrics "github.com/fesposito/go-ranger/ranger_metrics"
	"github.com/julienschmidt/httprouter"
)

var (
	logger ranger_logger.LoggerInterface
	rangerMetrics ranger_metrics.MetricsInterface
)

func init() {
	// we recommend to use ranger_logger (logrus + logstash hook)
	// if the connection fails we will warn and keep logging to stdout
	logger = ranger_logger.NewLoggerWithLogstashHook("tcp", "localhost:1234", "exampleApp")
	rangerMetrics = ranger_metrics.NewNewRelic("Your App Name", "<your-key-goes-here>....................", logger)
}

func main() {
	s := ranger_http.NewHTTPServer(logger).

		// you can add as many middlewares as  you want. they will be applied in the same order
		// sampleMiddlewar -> anotherSampleMiddleware -> ranger_http.RequestLog
		WithMiddleware(rangerMetrics.Middleware, sampleMiddleware, anotherSampleMiddleware, ranger_http.RequestLog).

		// with this we provide a default http 404 and 500 error.
		// see more on response_writer.go
		WithDefaultErrorRoute().

		// basic health check endpoints
		// /health/check/lb and /health/check
		// any struct sent as parameter here will be printed on key: value format (see Sprintf with "%+v")
		WithHealthCheckFor(nil).
		Build()

	// add some endpoints. based on "github.com/julienschmidt/httprouter"
	s.GET("/hello", helloEndpoint())

	addr := ":8080"

	// LoggerData is a map[string]interface{} struct
	logger.Info("Listening to address:", ranger_logger.LoggerData{"addr": addr})

	// decided to keep this startup process under control of the user, outside the go-ranger toolkit
	if err := http.ListenAndServe(addr, s.Start()); err != nil {
		logger.Error(err.Error(), nil)
	}
}

// some endpoints for demonstration purpose
func sampleMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger.Info("middleware", nil)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func anotherSampleMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger.Info("another middleware!", nil)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// just an example
func helloEndpoint() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode("Hello world")
	}
}
