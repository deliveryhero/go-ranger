package ranger_http

import (
	"net/http"

	"github.com/foodora/go-ranger/ranger_logger"
)

// PanicHandler is handling app panics gracefully
func PanicHandler(rw ResponseWriter) func(http.ResponseWriter, *http.Request, interface{}) {
	return func(w http.ResponseWriter, r *http.Request, ps interface{}) {
		logger.Error("Internal server error", ranger_logger.CreateFieldsFromRequest(r))
		rw.writeErrorResponse(w, http.StatusInternalServerError, "internalError", "Internal server error")
	}
}

// NotFoundHandler is handling 404s
func NotFoundHandler(rw ResponseWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warning("Resource not found", ranger_logger.CreateFieldsFromRequest(r))
		rw.writeErrorResponse(w, http.StatusNotFound, "resourceNotFound", "Resource not found")
	}
}
