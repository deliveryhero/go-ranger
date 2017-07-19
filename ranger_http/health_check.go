package ranger_http

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	Info   interface{} `json:"info"`
}

// HealthCheckHandler to check the service and external dependencies
func HealthCheckHandler(services []func() HealthCheckService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		mapServices := make(map[string]interface{})

		statusCode := http.StatusOK

		var service HealthCheckService
		for _, serviceFunc := range services {
			service = serviceFunc()

			mapServices[service.Name] = healthCheckServiceResponse{
				Status: service.Status,
				Info:   service.Info,
			}
			if service.Status == false {
				statusCode = http.StatusServiceUnavailable
			}
		}

		json.NewEncoder(w).Encode(
			healthCheckResponse{
				HTTPStatus: statusCode,
				Services:   mapServices,
			})
	}
}
