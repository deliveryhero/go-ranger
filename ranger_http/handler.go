package ranger_http

import (
	"fmt"
	"net/http"

	"github.com/foodora/go-ranger/ranger_logger"
)

// PanicHandler is handling app panics gracefully
func PanicHandler() func(http.ResponseWriter, *http.Request, interface{}) {
	return func(w http.ResponseWriter, r *http.Request, rcv interface{}) {
		loggerData := ranger_logger.CreateFieldsFromRequest(r)
		loggerData["error"] = rcv
		logger.Error(fmt.Sprintf("%s %s internalError", r.Method, r.RequestURI), loggerData)
		WriteErrorResponse(w, http.StatusInternalServerError, NewErrorResponseData("internalServerError", "Internal server error", ""))
	}
}

// NotFoundHandler is handling 404s
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warning(fmt.Sprintf("%s %s resourceNotFound", r.Method, r.RequestURI), ranger_logger.CreateFieldsFromRequest(r))
		WriteErrorResponse(w, http.StatusNotFound, NewErrorResponseData("resourceNotFound", "Resource not found", ""))
	}
}
