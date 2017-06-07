package http

import (
	"log"
	"net/http"
	"time"

	httprouter "github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"gopkg.in/throttled/throttled.v2"
	"gopkg.in/throttled/throttled.v2/store/memstore"
)

// Server ...
type Server struct {
	*httprouter.Router
	ResponseWriter
	middlewares []http.Handler
}

const (
	defaultResponseCacheTimeout = time.Duration(5) * time.Minute
)

// NewHTTPServer ...
func NewHTTPServer() *Server {
	responseWriter := ResponseWriter{}
	router := httprouter.New()

	return &Server{
		Router:         router,
		ResponseWriter: responseWriter,
	}
}

func (s *Server) WithDefaultErrorRoute() {
	s.PanicHandler = PanicHandler(s.ResponseWriter)
	s.NotFound = NotFoundHandler(s.ResponseWriter)
}

func (s *Server) WithHealthCheckFor(services ...interface{}) {
	s.GET("/health/check", HealthCheckHandler(services))
	s.GET("/health/check/lb", HealthCheckHandlerLB())
}

func (s *Server) WithMiddleware(middleware ...http.Handler) {
	s.middlewares = append(s.middlewares, middleware...)
}

func (s *Server) WithThrottle(handler *http.HandlerFunc) http.Handler {
	// @todo learn more about this memstore
	store, err := memstore.New(65536)
	if err != nil {
		log.Fatal(err)
	}

	quota := throttled.RateQuota{throttled.PerMin(20), 5}
	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		log.Fatal(err)
	}

	httpRateLimiter := throttled.HTTPRateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: true},
	}

	return httpRateLimiter.RateLimit(handler)
}

func (s *Server) Start(addr string) {
	// @todo append s.middlewares
	chain := alice.New().Then(s.Router)

	log.Print("Listening to address " + addr)
	log.Fatal(http.ListenAndServe(addr, chain))
}

// @todo add cache headers to response
