package ranger_http

import (
	"fmt"
	"net/http"

	"github.com/foodora/go-ranger/ranger_logger"
)

// PanicHandler is handling app panics gracefully
func PanicHandler(rw ResponseWriter) func(http.ResponseWriter, *http.Request, interface{}) {
	return func(w http.ResponseWriter, r *http.Request, ps interface{}) {
		logger.Error(fmt.Sprintf("%s %s internalError", r.Method, r.RequestURI), ranger_logger.CreateFieldsFromRequest(r))
		rw.writeErrorResponse(w, http.StatusInternalServerError, "internalError", "Internal server error")
	}
}

// NotFoundHandler is handling 404s
func NotFoundHandler(rw ResponseWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warning(fmt.Sprintf("%s %s resourceNotFound", r.Method, r.RequestURI), ranger_logger.CreateFieldsFromRequest(r))
		rw.writeErrorResponse(w, http.StatusNotFound, "resourceNotFound", "Resource not found")
	}
}
