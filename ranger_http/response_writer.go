package ranger_http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse struct
type ErrorResponse struct {
	Status int `json:"status"`
	Data *ErrorResponseData `json:"data"`
}

// ErrorResponseData struct
type ErrorResponseData struct {
	ErrorCode       string `json:"exception_type"`
	Message         string `json:"message"`
	MoreInformation	string `json:"more_information"`
}

// NewErrorResponseData ...
func NewErrorResponseData(errorCode string, message string, moreInformation string) *ErrorResponseData {
	return &ErrorResponseData{
		ErrorCode: errorCode,
		Message: message,
		MoreInformation: moreInformation,
	}
}

// WriteErrorResponse Write status code and body for an error response
func WriteErrorResponse(rw http.ResponseWriter, statusCode int, erd *ErrorResponseData) {
	rw.WriteHeader(statusCode)

	json.NewEncoder(rw).Encode(ErrorResponse{
		Status: statusCode,
		Data:   erd,
	})
}
