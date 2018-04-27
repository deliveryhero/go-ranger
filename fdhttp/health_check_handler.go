package fdhttp

import (
	"context"
	"net/http"
	"os"
	"time"
)

// HealthCheckService is the interface that your service need to provide to
// be able to health check
type HealthCheckService interface {
	HealthCheck() (interface{}, error)
}

// HealthCheckResponse is the main json response from healthcheck endpoint
type HealthCheckResponse struct {
	Version struct {
		Tag    string `json:"tag"`
		Commit string `json:"commit"`
	} `json:"version"`
	Status   bool                                  `json:"status"`
	Elapsed  time.Duration                         `json:"elapsed"`
	Hostname string                                `json:"hostname"`
	Checks   map[string]HealthCheckServiceResponse `json:"checks,omitempty"`
}

// HealthCheckServiceResponse is the return of each service that can provide
// details
type HealthCheckServiceResponse struct {
	Status  bool          `json:"status"`
	Elapsed time.Duration `json:"elapsed"`
	Detail  interface{}   `json:"detail,omitempty"`
	Error   interface{}   `json:"error,omitempty"`
}

// HealthCheckServiceError can be returned as a error and Detail()
// will populate Error response
type HealthCheckServiceError interface {
	error
	Detail() interface{}
}

var _ Handler = &HealthCheckHandler{}

// DefaultHealthCheckURL is the urlto access your health check.
// Note you can specify a Prefix (inside of HealthCheckHandler object)
// and you also can check a specific service using: Prefix + DefaultHealthCheckURL + "/<service-name>"
var DefaultHealthCheckURL = "/health/check"

// HealthCheckHandler is a valid http handler to export app version and it also checks
// service registred.
type HealthCheckHandler struct {
	// Prefix will be prefix the fdhttp.DefaultHealthCheckURL.
	Prefix string

	tag      string
	commit   string
	hostname string
	services map[string]HealthCheckService
}

// NewHealthCheckHandler create a new healthcheck handler
func NewHealthCheckHandler(tag, commit string) *HealthCheckHandler {
	hostname, _ := os.Hostname()

	return &HealthCheckHandler{
		tag:      tag,
		commit:   commit,
		hostname: hostname,
		services: make(map[string]HealthCheckService),
	}
}

func (h *HealthCheckHandler) Init(r *Router) {
	r.GET(h.Prefix+DefaultHealthCheckURL, h.Get)
	r.GET(h.Prefix+DefaultHealthCheckURL+"/:service", h.Get)
}

// Register a new healthcheck Service
func (h *HealthCheckHandler) Register(name string, s HealthCheckService) {
	h.services[name] = s
}

func (h *HealthCheckHandler) Get(ctx context.Context) (int, interface{}) {
	started := time.Now()

	serviceParam := RouteParam(ctx, "service")

	statusCode := http.StatusOK
	resp := HealthCheckResponse{
		Status:   true,
		Hostname: h.hostname,
		Checks:   make(map[string]HealthCheckServiceResponse),
	}
	resp.Version.Tag = h.tag
	resp.Version.Commit = h.commit

	for name, s := range h.services {
		if serviceParam != "" && name != serviceParam {
			continue
		}

		started := time.Now()
		serviceCheck := HealthCheckServiceResponse{
			Status: true,
		}

		detail, err := s.HealthCheck()
		if err != nil {
			statusCode = http.StatusServiceUnavailable
			resp.Status = false
			serviceCheck.Status = false

			if hcErr, ok := err.(HealthCheckServiceError); ok {
				serviceCheck.Error = hcErr.Detail()
			} else {
				serviceCheck.Error = err.Error()
			}

		} else {
			serviceCheck.Detail = detail
		}

		serviceCheck.Elapsed = time.Since(started) / time.Millisecond
		resp.Checks[name] = serviceCheck
	}

	resp.Elapsed = time.Since(started) / time.Millisecond

	return statusCode, resp
}
