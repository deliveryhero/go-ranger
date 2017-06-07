package http

import (
	"fmt"
	"log"
	"net/http"
	"time"

	httprouter "github.com/julienschmidt/httprouter"
)

// @todo middleware

// Server ...
type Server struct {
	*httprouter.Router
	ResponseWriter
}

const (
	defaultResponseCacheTimeout = time.Duration(5) * time.Minute
)

// NewHTTPServer ...
func NewHTTPServer() *Server {
	responseWriter := ResponseWriter{}
	router := httprouter.New()

	router.PanicHandler = PanicHandler(responseWriter)
	router.NotFound = NotFoundHandler(responseWriter)

	router.GET("/health/check", HealthCheckHandler())
	router.GET("/health/check/lb", HealthCheckHandlerLB())

	return &Server{
		Router:         router,
		ResponseWriter: responseWriter,
	}
}

func (s *Server) addMiddleware(middleware http.Handler) http.Handler {
	return nil
}

func (s *Server) start(addr string) {
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
