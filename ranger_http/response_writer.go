package ranger_http

import (
	"encoding/json"
	"net/http"
)

type StatusResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w *StatusResponseWriter) Status() int {
	return w.status
}

func (w *StatusResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *StatusResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w *StatusResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// ErrorResponse struct
type ErrorResponse struct {
	Status int                `json:"status"`
	Data   *ErrorResponseData `json:"data"`
}

// ErrorResponseData struct
type ErrorResponseData struct {
	ErrorCode       string `json:"exception_type"`
	Message         string `json:"message"`
	MoreInformation string `json:"more_information"`
}

// NewErrorResponseData ...
func NewErrorResponseData(errorCode, message, moreInformation string) *ErrorResponseData {
	return &ErrorResponseData{
		ErrorCode:       errorCode,
		Message:         message,
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
