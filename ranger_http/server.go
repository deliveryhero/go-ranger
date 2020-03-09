package ranger_http

import (
	"net/http"
	"time"

	"github.com/foodora/go-ranger/ranger_logger"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/throttled/throttled"
	"github.com/throttled/throttled/store/memstore"
)

// Server ...
type Server struct {
	*httprouter.Router
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
	router := httprouter.New()

	return &Server{
		Router: router,
	}
}

// WithDefaultErrorRoute ...
func (s *Server) WithDefaultErrorRoute() *Server {
	s.PanicHandler = PanicHandler()
	s.NotFound = NotFoundHandler()
	return s
}

// WithMiddleware ...
func (s *Server) WithMiddleware(middlewares ...func(http.Handler) http.Handler) *Server {
	for _, v := range middlewares {
		s.middlewares = append(s.middlewares, v)
	}
	return s
}

// SetThrottle ...
func (s *Server) SetThrottle(handler *http.HandlerFunc) http.Handler {
	// @todo learn more about this memstore
	store, err := memstore.New(65536)
	if err != nil {
		logger.Error(err.Error(), nil)
	}

	quota := throttled.RateQuota{MaxRate: throttled.PerMin(20), MaxBurst: 5}
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

// Start ...
func (s *Server) Start() http.Handler {
	chain := alice.New(s.middlewares...)
	return chain.Then(s.Router)
}

// @todo add cache headers to response
