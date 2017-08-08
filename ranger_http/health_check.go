package ranger_http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

//HealthCheckService
type HealthCheckService struct {
	Name   string
	Status bool
	Info   interface{}
}

//HealthCheckConfiguration represents a configuration service
type healthCheckConfiguration struct {
	Prefix   string
	Services []func() HealthCheckService
}

//NewHealthCheckConfiguration
func NewHealthCheckConfiguration(services ...func() HealthCheckService) healthCheckConfiguration {
	return healthCheckConfiguration{
		Services: services,
	}
}

//WithPrefix
func (configuration healthCheckConfiguration) WithPrefix(prefix string) healthCheckConfiguration {
	configuration.Prefix = prefix

	return configuration
}

type healthCheckResponse struct {
	HTTPStatus int                    `json:"http-status"`
	Time 	   float64                `json:"time"`
	Services   map[string]interface{} `json:"checks"`
}

//WithHealthCheckFor ...
func (s Server) WithHealthCheckFor(configuration healthCheckConfiguration) Server {
	s.GET(fmt.Sprintf("%s/health/check", configuration.Prefix), HealthCheckHandler(configuration.Services))
	s.GET(fmt.Sprintf("%s/health/check/lb", configuration.Prefix), HealthCheckHandlerLB())
	return s
}

// HealthCheckHandlerLB to check if the webserver is up
func HealthCheckHandlerLB() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
}

type healthCheckServiceResponse struct {
	Status bool        `json:"status"`
	Time   float64     `json:"time"`
	Info   interface{} `json:"info"`
}

// HealthCheckHandler to check the service and external dependencies
func HealthCheckHandler(services []func() HealthCheckService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		mapServices := make(map[string]interface{})

		statusCode := http.StatusOK
		sAll := time.Now()

		var service HealthCheckService
		for _, serviceFunc := range services {
			if serviceFunc != nil {
				s := time.Now()
				service = serviceFunc()

				mapServices[service.Name] = healthCheckServiceResponse{
					Status: service.Status,
					Time: ElapsedTimeSince(s),
					Info:   service.Info,
				}
				if service.Status == false {
					statusCode = http.StatusServiceUnavailable
				}
			}
		}

		json.NewEncoder(w).Encode(
			healthCheckResponse{
				HTTPStatus: statusCode,
				Time: ElapsedTimeSince(sAll),
				Services:   mapServices,
			})
	}
}

//ElapsedTimeSince calculates the elapsed time from a given start
func ElapsedTimeSince(s time.Time) float64 {
	return float64(time.Since(s)) / float64(time.Second)
}
