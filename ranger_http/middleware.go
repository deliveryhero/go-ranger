package ranger_http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/foodora/go-ranger/ranger_logger"
)

//MiddlewareInterface
type MiddlewareInterface interface {
	Middleware(next http.Handler) http.Handler
}

//NewRequestLogger - RequestLogger Constructor
func NewRequestLogger() *RequestLogger {
	return &RequestLogger{}
}

//RequestLogger struct
type RequestLogger struct {
}

//Middleware - RequestLogger middleware
func (requestLogger *RequestLogger) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			var message bytes.Buffer
			message.WriteString(r.Method)
			message.WriteString(" ")
			message.WriteString(r.RequestURI)

			logger.Info(
				message.String(),
				ranger_logger.LoggerData{
					"method": r.Method,
					"URI":    r.RequestURI,
					"time":   time.Since(start),
				},
			)
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
