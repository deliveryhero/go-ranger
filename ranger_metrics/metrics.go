package ranger_metrics

import "net/http"

//MetricsInterface
type MetricsInterface interface {
	Middleware(next http.Handler) http.Handler
}
