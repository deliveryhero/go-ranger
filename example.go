package main

import (
	"net/http"
	"os"

	"time"

	ranger_http "github.com/foodora/go-ranger/ranger_http"
	ranger_logger "github.com/foodora/go-ranger/ranger_logger"
	ranger_metrics "github.com/foodora/go-ranger/ranger_metrics"
	ranger_os "github.com/foodora/go-ranger/ranger_os"
	"github.com/julienschmidt/httprouter"
)

var (
	logger   ranger_logger.LoggerInterface
	logstash ranger_logger.Hook
	slack    ranger_logger.Hook
	metrics  *ranger_metrics.NewRelic
)

func init() {
	// load env vars if present
	ranger_os.ExportEnvVars(".env")

	// our logging system accepts power-ups called hooks
	// you can use all logrus hooks. we provide some useful with go-ranger

	// logstash hook
	logstash = ranger_logger.NewLogstashHook(
		"tcp",
		"localhost:1234",
		// important to use this formatter to have correct timestamps
		ranger_logger.GetJSONFormatter(),
	)

	// slack hook
	slack = ranger_logger.NewSlackHook(
		"#my-channel",
		"https://hooks.slack.com/services/T00/B00/absfmzyy",
		"debug",
	)

	logger = ranger_logger.NewLogger(
		os.Stdout,
		ranger_logger.LoggerData{"environment": "development"},
		&ranger_logger.JSONFormatter{},
		"debug",
		logstash,
		slack,
	)

	// currenttly we only support new relic APM integration
	metrics = ranger_metrics.NewNewRelic("Your App Name", "<your-key-goes-here>....................", logger)
}

func main() {
	s := ranger_http.NewHTTPServer(logger).

		// you can add as many middlewares as you want. they will be applied in the same order
		// sampleMiddlewar -> anotherSampleMiddleware -> ranger_http.RequestLog
		WithMiddleware(
			metrics.Middleware,
			sampleMiddleware,
			anotherSampleMiddleware,
			ranger_http.LoggerMiddleware,
		).

		// with this we provide a default http 404 and 500 error.
		// see more on response_writer.go
		WithDefaultErrorRoute().

		// basic health check endpoints
		// /health/check/lb and /health/check
		// any instance of `func() ranger_http.HealthCheckService` sent as parameter of the configuration will be converted to JSON and printed
		// if necessary, you also can add a prefix to the endpoints with WithPrefix("/prefix")
		//     ex: WithHealthCheckFor(ranger_http.NewHealthCheckConfiguration(versionHealthCheck()).WithPrefix("/prefix"))
		WithHealthCheckFor(ranger_http.NewHealthCheckConfiguration(versionHealthCheck(), etcdHealthCheck()))

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

		// calling a external service
		apiClient := ranger_http.NewAPIClient(5)

		// starting cross application transaction
		// more info here "Making HTTP requests" https://docs.newrelic.com/docs/agents/go-agent/features/cross-application-tracing-go
		txn := metrics.Application.StartTransaction("cross-application-transaction-example", nil, nil)
		defer txn.End()
		metrics.UseNewRoundTripper(txn, apiClient.Client)

		body, err := apiClient.GetContentByURL("GET", "http://www.mocky.io/v2/5185415ba171ea3a00704eed", nil)
		if err != nil {
			logger.Error("Could not make mock call", ranger_logger.LoggerData{"error": err})
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		// normally we use a encoder to return a struct
		//json.NewEncoder(w).Encode("Hi gopher!")

		// but in this case we can simple forward the returned json
		w.Write(body)
	}
}

func versionHealthCheck() func() ranger_http.HealthCheckService {
	type versionHealthCheck struct {
		Tag    string `json:"tag"`
		Commit string `json:"commit"`
	}

	return func() ranger_http.HealthCheckService {
		return ranger_http.HealthCheckService{
			Name:   "version",
			Status: true,
			Info: versionHealthCheck{
				Tag:    "0.0.0",
				Commit: "30ac4383d0260f517d7f171de244fa46c1879c67",
			},
		}
	}
}

func etcdHealthCheck() func() ranger_http.HealthCheckService {
	type etcdHealthCheck struct {
		ResponseTime int `json:"response_time"`
	}

	return func() ranger_http.HealthCheckService {
		//some logic here to get etcd response time
		var crazyLogic int
		crazyLogic = int(time.Now().Unix() % 10)

		return ranger_http.HealthCheckService{
			Name:   "etcd",
			Status: true,
			Info: etcdHealthCheck{
				ResponseTime: crazyLogic,
			},
		}
	}
}
