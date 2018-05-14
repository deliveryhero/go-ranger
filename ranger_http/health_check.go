package ranger_http

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
	Version  healthCheckVersion
}

type healthCheckVersion struct {
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

//NewHealthCheckConfiguration
func NewHealthCheckConfiguration(services ...func() HealthCheckService) healthCheckConfiguration {
	return healthCheckConfiguration{
		Services: services,
		Version:  healthCheckVersion{"n/a", "n/a"},
	}
}

//WithPrefix
func (configuration healthCheckConfiguration) WithPrefix(prefix string) healthCheckConfiguration {
	configuration.Prefix = prefix

	return configuration
}

// WithVersion expects the path to a json file for healthCheckVersion
func (configuration healthCheckConfiguration) WithVersion(versionPath string) healthCheckConfiguration {

	version := healthCheckVersion{}
	configuration.Version = version

	absPath, _ := filepath.Abs(versionPath)
	fileBytes, err := ioutil.ReadFile(absPath)

	if err != nil {
		logger.Warning("Unable to load file "+versionPath, nil)
	}
	err = json.Unmarshal(fileBytes, &version)
	if err != nil {
		logger.Warning("Error parsing version file "+versionPath, nil)
	}

	configuration.Version = version

	return configuration
}

type healthCheckResponse struct {
	HTTPStatus int                    `json:"http-status"`
	Version    healthCheckVersion     `json:"version"`
	Time       float64                `json:"time"`
	Status     bool                   `json:"status"`
	Host       string                 `json:"hostname"`
	Services   map[string]interface{} `json:"checks"`
}

//WithHealthCheckFor ...
func (s Server) WithHealthCheckFor(configuration healthCheckConfiguration) Server {
	s.GET(fmt.Sprintf("%s/health/check", configuration.Prefix), HealthCheckHandler(configuration))
	s.GET(fmt.Sprintf("%s/health/check/lb", configuration.Prefix), HealthCheckHandlerLB())
	return s
}

// HealthCheckHandlerLB to check if the webserver is up
func HealthCheckHandlerLB() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
		w.WriteHeader(http.StatusOK)
	}
}

type healthCheckServiceResponse struct {
	Status bool        `json:"status"`
	Time   float64     `json:"time"`
	Info   interface{} `json:"info"`
}

// HealthCheckHandler to check the service and external dependencies
func HealthCheckHandler(configuration healthCheckConfiguration) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Cache-Control", "no-cache, private, max-age=0")

		services := configuration.Services
		hostname, _ := os.Hostname()

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
					Time:   ElapsedTimeSince(s),
					Info:   service.Info,
				}
				if service.Status == false {
					statusCode = http.StatusServiceUnavailable
				}
			}
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(
			healthCheckResponse{
				HTTPStatus: statusCode,
				Version:    configuration.Version,
				Time:       ElapsedTimeSince(sAll),
				Status:     service.Status,
				Host:       hostname,
				Services:   mapServices,
			})
	}
}

//ElapsedTimeSince calculates the elapsed time from a given start
func ElapsedTimeSince(s time.Time) float64 {
	return time.Since(s).Seconds()
}
