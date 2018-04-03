package fdhttp

import (
	"fmt"
	"net/http"
)

// NewNotFoundHandler return a handler, used by default, to deal
// with routes not found
func NewNotFoundHandler() http.HandlerFunc {
	return notFoundHandler
}

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	ResponseJSON(w, http.StatusNotFound, &ResponseError{
		Code:    "not_found",
		Message: fmt.Sprintf("URL '%s' was not found", req.URL.String()),
	})
}
