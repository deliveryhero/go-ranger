package fdhttp

import (
	"fmt"
	"net/http"
)

// NewMethodNotAllowedHandler return a handler, used by default, to deal
// with method not allowed
func NewMethodNotAllowedHandler() http.HandlerFunc {
	return methodNotAllowedHandler
}

func methodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	ResponseJSON(w, http.StatusMethodNotAllowed, &ResponseError{
		Code:    "method_not_allowed",
		Message: fmt.Sprintf("Method '%s' is not allowed to access '%s'", req.Method, req.URL.String()),
	})
}
