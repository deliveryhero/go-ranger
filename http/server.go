package http

import (
	"fmt"
	"log"
	"net/http"
	"time"

	httprouter "github.com/julienschmidt/httprouter"
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

func (s *Server) withDefaultRoutes() {
	s.Router.PanicHandler = PanicHandler(s.ResponseWriter)
	s.Router.NotFound = NotFoundHandler(s.ResponseWriter)
}

func (s *Server) withHealthCheck(services ...interface{}) {
	s.Router.GET("/health/check", HealthCheckHandler(services))
	s.Router.GET("/health/check/lb", HealthCheckHandlerLB())
}

func (s *Server) withMiddleware(middleware ...http.Handler) {
	s.middlewares = append(s.middlewares, middleware...)
}

func (s *Server) withThrotle() {
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

	log.Print(httpRateLimiter)
	// @todo s.middlewares = append(s.middlewares, httpRateLimiter.RateLimit(h))

}

func (s *Server) start(addr string) {

	//chain := alice.New(httpRateLimiter.RateLimiter).Then(myHandler)
	//http.ListenAndServe(":8000", chain)

	http.ListenAndServe(":8000", s.Router)
	log.Print("Listening to address " + addr)
	log.Fatal(http.ListenAndServe(addr, s.Router))
}

func (s *Server) addCacheToHeader(w http.ResponseWriter) http.ResponseWriter {
	cacheSince := time.Now().Format(http.TimeFormat)
	cacheUntil := time.Now().Add(defaultResponseCacheTimeout).Format(http.TimeFormat)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%.0f", defaultResponseCacheTimeout.Seconds()))
	w.Header().Set("Date", cacheSince)
	w.Header().Set("Expires", cacheUntil)
	return w
}
