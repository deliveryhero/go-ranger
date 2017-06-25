package ranger_http

import (
	"net/http"
	"time"

	"github.com/foodora/go-ranger/ranger_logger"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"gopkg.in/throttled/throttled.v2"
	"gopkg.in/throttled/throttled.v2/store/memstore"
)

// Server ...
type Server struct {
	*httprouter.Router
	ResponseWriter
	middlewares []alice.Constructor
}

const (
	defaultResponseCacheTimeout = time.Duration(5) * time.Minute
)

var (
	logger ranger_logger.LoggerInterface
)

// NewHTTPServer ...
func NewHTTPServer(l ranger_logger.LoggerInterface) *Server {
	logger = l
	responseWriter := ResponseWriter{}
	router := httprouter.New()

	return &Server{
		Router:         router,
		ResponseWriter: responseWriter,
	}
}

func (s Server) WithDefaultErrorRoute() Server {
	s.PanicHandler = PanicHandler(s.ResponseWriter)
	s.NotFound = NotFoundHandler(s.ResponseWriter)
	return s
}

func (s Server) WithHealthCheckFor(healthCheckPath string, healthCheckLbPath string, services ...interface{}) Server {
	s.GET(healthCheckPath, HealthCheckHandler(services))
	s.GET(healthCheckLbPath, HealthCheckHandlerLB())
	return s
}

func (s Server) WithMiddleware(middlewares ...func(http.Handler) http.Handler) Server {
	for _, v := range middlewares {
		s.middlewares = append(s.middlewares, v)
	}
	return s
}

func (s Server) SetThrottle(handler *http.HandlerFunc) http.Handler {
	// @todo learn more about this memstore
	store, err := memstore.New(65536)
	if err != nil {
		logger.Error(err.Error(), nil)
	}

	quota := throttled.RateQuota{throttled.PerMin(20), 5}
	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		logger.Error(err.Error(), nil)
	}

	httpRateLimiter := throttled.HTTPRateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: true},
	}

	return httpRateLimiter.RateLimit(handler)
}

func (s Server) Build() Server {
	return s
}

func (s Server) Start() http.Handler {
	chain := alice.New(s.middlewares...)
	return chain.Then(s.Router)
}

// @todo add cache headers to response
