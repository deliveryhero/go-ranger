package ranger_http

import (
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
			loggerData := ranger_logger.CreateFieldsFromRequest(r)
			loggerData["response_time"] = time.Since(start).Seconds()

			logger.Debug(
				"ranger_logger.LoggerMiddleware",
				loggerData,
			)
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
