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

// LoggerMiddleware ...
func LoggerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := StatusResponseWriter{http.StatusOK, w}

		defer func() {
			var message bytes.Buffer
			message.WriteString(r.Method)
			message.WriteString(" ")
			message.WriteString(r.RequestURI)

			logger.Debug(
				message.String(),
				ranger_logger.LoggerData{
					"method": r.Method,
					"uri":    r.RequestURI,
					"status": sr.Status(),
					"time":   time.Since(start),
				},
			)
		}()

		next.ServeHTTP(&sr, r)
	}

	return http.HandlerFunc(fn)
}
