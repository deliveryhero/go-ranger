package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// ResponseWriter struct
type ResponseWriter struct {
}

// HealthCheckResponse represents all checked services
type HealthCheckResponse struct {
	HTTPStatus int    `json:"http_status"`
	Version    string `json:"version"`
	Services   string `json:"services"`
}

// ErrorResponse struct
type ErrorResponse struct {
	Status int `json:"status"`
	Data   struct {
		ErrorCode       string `json:"exception_type"`
		Message         string `json:"message"`
		Details         string `json:"developer_message"`
		MoreInformation string `json:"more_information"`
	} `json:"data"`
}

func (writer *ResponseWriter) writeErrorResponse(rw http.ResponseWriter, statusCode int, errorCode string, message string) {
	rw.WriteHeader(statusCode)

	json.NewEncoder(rw).Encode(ErrorResponse{
		Status: statusCode,
		Data: struct {
			ErrorCode       string `json:"exception_type"`
			Message         string `json:"message"`
			Details         string `json:"developer_message"`
			MoreInformation string `json:"more_information"`
		}{
			ErrorCode:       errorCode,
			Message:         message,
			Details:         "",
			MoreInformation: "null",
		},
	})

	log.Println(errorCode + " - " + message)
}
