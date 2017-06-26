package ranger_http

import (
	"net/http"
)

// PanicHandler is handling app panics gracefully
func PanicHandler(rw ResponseWriter) func(http.ResponseWriter, *http.Request, interface{}) {
	return func(w http.ResponseWriter, r *http.Request, ps interface{}) {
		rw.writeErrorResponse(w, http.StatusInternalServerError, "internalError", "Internal server error")
	}
}

// NotFoundHandler is handling 404s
func NotFoundHandler(rw ResponseWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw.writeErrorResponse(w, http.StatusNotFound, "resourceNotFound", "Resource not found")
	}
}
