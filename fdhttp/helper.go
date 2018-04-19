package fdhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ResponseError is an error struct that can be used to return error as json
type ResponseError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

// Error implements error interface
//  func ReturnError() error {
//      return &ResponseError{Code: "not_found", Message: "invalid id"}
//  }
func (err *ResponseError) Error() string {
	return fmt.Sprintf("%s: %s", err.Code, err.Message)
}

// ResponseJSON respond as a json object.
func ResponseJSON(w http.ResponseWriter, statusCode int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if resp == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

// Un can be called with defer passing Lock() function as parameter.
//  defer fdhttp.Un(Lock(&m))
func Un(f func()) {
	f()
}

// Lock m and return a function to unlock.
//  unlock := fdhttp.Lock(&m)
//  // your code here
//  unlock()
func Lock(m sync.Locker) func() {
	m.Lock()
	return func() { m.Unlock() }
}
