package ranger_http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//HealthCheckService
type HealthCheckService struct {
	name string
	data interface{}
}

//HealthCheckConfiguration represents a configuration service
type healthCheckConfiguration struct {
	Prefix   string
	Services []HealthCheckService
}

//NewHealthCheckConfiguration
func NewHealthCheckConfiguration(services ...HealthCheckService) healthCheckConfiguration {
	return healthCheckConfiguration{
		Services: services,
	}
}

//WithPrefix
func (configuration healthCheckConfiguration) WithPrefix(prefix string) healthCheckConfiguration {
	configuration.Prefix = prefix

	return configuration
}

//NewHealthCheckService constructor
func NewHealthCheckService(name string, data interface{}) HealthCheckService {
	return HealthCheckService{
		name: name,
		data: data,
	}
}

type healthCheckResponse struct {
	HTTPStatus int                    `json:"http_status"`
	Services   map[string]interface{} `json:"services"`
}

//WithHealthCheckFor
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

// HealthCheckHandler to check the service and external dependencies
func HealthCheckHandler(services []HealthCheckService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		mapServices := make(map[string]interface{})

		for _, service := range services {
			mapServices[service.name] = service.data
		}

		json.NewEncoder(w).Encode(
			healthCheckResponse{
				HTTPStatus: http.StatusOK,
				Services:   mapServices,
			})
	}
}
