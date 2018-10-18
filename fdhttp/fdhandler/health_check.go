package fdhandler

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
)

// HealthChecker is the interface that your service need to provide to
// be able to check its health.
type HealthChecker interface {
	HealthCheck(context.Context) (interface{}, error)
}

// HealthCheckResponse is the main json response from healthcheck endpoint
type HealthCheckResponse struct {
	Version struct {
		Tag    string `json:"tag"`
		Commit string `json:"commit"`
	} `json:"version"`
	Status   bool                                   `json:"status"`
	Elapsed  time.Duration                          `json:"elapsed"`
	Hostname string                                 `json:"hostname"`
	Checks   map[string]*HealthCheckServiceResponse `json:"checks,omitempty"`
	System   struct {
		Version         string `json:"version"`
		NumCPU          int    `json:"num_cpu"`
		NumGoroutines   int    `json:"num_goroutines"`
		NumHeapObjects  uint64 `json:"num_heap_objects"`
		TotalAllocBytes uint64 `json:"total_alloc_bytes"`
		AllocBytes      uint64 `json:"alloc_bytes"`
	} `json:"system,omitempty"`
}

// HealthCheckServiceResponse is the return of each service that can provide
// details
type HealthCheckServiceResponse struct {
	Status  bool          `json:"status"`
	Elapsed time.Duration `json:"elapsed"`
	Detail  interface{}   `json:"detail,omitempty"`
	Error   interface{}   `json:"error,omitempty"`
}

var _ fdhttp.Handler = &HealthCheck{}

// HealthCheckURL is the urlto access your health check.
// Note you can specify a Prefix (inside of HealthCheck object)
// and you also can check a specific service using: Prefix + HealthCheckURL + "/<service-name>"
var HealthCheckURL = "/health/check"

// HealthCheckServiceTimeout is the time that your service need to return otherwise
var HealthCheckServiceTimeout = 1 * time.Second

// HealthCheck is a valid http handler to export app version and it also checks
// service registred.
type HealthCheck struct {
	// Prefix will be prefix the fdhttp.HealthCheckURL.
	Prefix string
	// ServiceTimeout is max duration that each service should return
	// before we get a timeout.
	ServiceTimeout time.Duration

	tag      string
	commit   string
	hostname string
	services map[string]HealthChecker
}

// NewHealthCheck create a new healthcheck handler
func NewHealthCheck(tag, commit string) *HealthCheck {
	hostname, _ := os.Hostname()

	return &HealthCheck{
		tag:            tag,
		commit:         commit,
		hostname:       hostname,
		services:       make(map[string]HealthChecker),
		ServiceTimeout: HealthCheckServiceTimeout,
	}
}

// Init will be called by fdhttp.Router to register fdhandler.HealthCheckURL
// into it.
func (h *HealthCheck) Init(r *fdhttp.Router) {
	r.GET(h.Prefix+HealthCheckURL, h.Get)
	r.GET(h.Prefix+HealthCheckURL+"/:service", h.Get)
}

// Register a new healthcheck Service.
func (h *HealthCheck) Register(name string, s HealthChecker) {
	h.services[name] = s
}

// Handler will return an http.Handler that you can register in your
// http.Server.
func (h *HealthCheck) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		statusCode, resp := h.Get(req.Context())
		fdhttp.ResponseJSON(w, statusCode, resp)
	})
}

func (h *HealthCheck) newResponse() *HealthCheckResponse {
	resp := &HealthCheckResponse{
		Status:   true,
		Hostname: h.hostname,
		Checks:   make(map[string]*HealthCheckServiceResponse),
	}

	resp.Version.Tag = h.tag
	resp.Version.Commit = h.commit

	resp.System.Version = runtime.Version()
	resp.System.NumCPU = runtime.NumCPU()
	resp.System.NumGoroutines = runtime.NumGoroutine()

	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	resp.System.AllocBytes = mem.HeapAlloc
	resp.System.TotalAllocBytes = mem.TotalAlloc
	resp.System.NumHeapObjects = mem.HeapObjects

	return resp
}

// Get is a fdhttp.EndpointFunc that will be registred in the fdhttp.Router.
// If you don't plan to use fdhttp.Router, check HealthCheck.Handler()
func (h *HealthCheck) Get(ctx context.Context) (int, interface{}) {
	started := time.Now()
	svcParam := fdhttp.RouteParam(ctx, "service")

	var statusCode int32 = http.StatusOK

	resp := h.newResponse()

	var wg sync.WaitGroup
	for name, svc := range h.services {
		if svcParam != "" && name != svcParam {
			continue
		}

		svcCheck := &HealthCheckServiceResponse{
			Status: true,
		}

		resp.Checks[name] = svcCheck

		wg.Add(1)
		go func(svc HealthChecker, check *HealthCheckServiceResponse) {
			defer wg.Done()
			started := time.Now()

			timeoutCtx, cancel := context.WithTimeout(ctx, h.ServiceTimeout)
			go func() {
				detail, err := svc.HealthCheck(timeoutCtx)

				if timeoutCtx.Err() != nil {
					// context has timed out
					return
				}

				// cancel the context
				cancel()

				if err != nil {
					atomic.CompareAndSwapInt32(&statusCode, http.StatusOK, http.StatusServiceUnavailable)
					resp.Status = false
					check.Status = false
					check.Error = err.Error()
				}
				check.Detail = detail
			}()

			<-timeoutCtx.Done()

			err := timeoutCtx.Err()
			if err == context.DeadlineExceeded {
				atomic.CompareAndSwapInt32(&statusCode, http.StatusOK, http.StatusRequestTimeout)
				resp.Status = false
				check.Status = false
				check.Error = err.Error()
			}

			check.Elapsed = time.Since(started) / time.Millisecond
		}(svc, svcCheck)
	}

	wg.Wait()

	resp.Elapsed = time.Since(started) / time.Millisecond

	return int(statusCode), resp
}
