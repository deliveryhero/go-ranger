package ranger_http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
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

// HealthCheckHandlerLB to check if the webserver is up
func HealthCheckHandlerLB() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
}

// HealthCheckHandler to check the service and external dependencies
func HealthCheckHandler(services ...interface{}) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// @todo receive this as param
		version := "0.0.0"

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(
			HealthCheckResponse{
				HTTPStatus: http.StatusOK,
				Version:    version,
				Services:   fmt.Sprintf("%+v", services...),
			})
	}
}
