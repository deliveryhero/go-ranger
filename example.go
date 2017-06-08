package main

import (
	"encoding/json"
	"net/http"

	ranger_http "github.com/fesposito/go-ranger/ranger_http"
	ranger_logger "github.com/fesposito/go-ranger/ranger_logger"
	"github.com/julienschmidt/httprouter"
)

var (
	logger ranger_logger.LoggerInterface
)

func init() {
	// we recommend to use logrus + logstash hook
	// if the connection fails we will warn and keep logging to stdout
	logger = ranger_logger.NewLoggerWithLogstashHook("tcp", "localhost:1234", "exampleApp", nil)
}

func main() {
	s := ranger_http.NewHTTPServer(logger)

	// you can add as many middlewares as  you want. they will be applied in the same order
	s.WithMiddleware(sampleMiddleware, anotherSampleMiddleware, ranger_http.RequestLog)

	// with this we provide a default http 404 and 500 error
	s.WithDefaultErrorRoute()

	// basic health check endpoints
	// /health/check/lb and /health/check
	// any struct passed as parameter here will be printed on key: value format (see Sprintf with "%+v")
	s.WithHealthCheckFor(nil)

	// add more endpoints. based on "github.com/julienschmidt/httprouter"
	s.GET("/hello", helloEndpoint())

	// defines the app port
	s.Start(":8080")
}

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

func helloEndpoint() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode("Hello world")
	}
}
