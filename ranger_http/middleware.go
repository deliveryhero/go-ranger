package ranger_http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/foodora/go-ranger/ranger_logger"
	"net/http/httputil"
)

//MiddlewareInterface
type MiddlewareInterface interface {
	Middleware(next http.Handler) http.Handler
}

// LoggerMiddleware ...
func LoggerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			var message bytes.Buffer
			message.WriteString(r.Method)
			message.WriteString(" ")
			message.WriteString(r.RequestURI)

			logger.Debug(
				message.String(),
				ranger_logger.LoggerData{
					"method":  r.Method,
					"URI":     r.RequestURI,
					"time":    time.Since(start),
					"request": getRequestDump(r),
				},
			)
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

//getRequestDump - Get a pretty print request or empty string
func getRequestDump(r *http.Request) string {
	d, err := httputil.DumpRequest(r, true)
	if err != nil {
		return ""
	}

	return string(d)
}
