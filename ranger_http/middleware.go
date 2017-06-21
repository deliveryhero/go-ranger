package ranger_http

import (
	"net/http"
	"time"

	ranger_logger "github.com/foodora/go-ranger/ranger_logger"
)

func RequestLog(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// @todo add logrus CreateFieldsFromRequest call

		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		logger.Info("RequestLog", ranger_logger.LoggerData{"method": r.Method, "url": r.URL.String(), "t": t2.Sub(t1)})
	}
	return http.HandlerFunc(fn)
}
