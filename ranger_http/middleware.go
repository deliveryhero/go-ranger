package ranger_http

import (
	"bytes"
	"io/ioutil"
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
					"body": getRequestBody(r),
				},
			)
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

//getRequestBody - Get a pretty print request or empty string
func getRequestBody(r *http.Request) string {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	return string(body)
}
