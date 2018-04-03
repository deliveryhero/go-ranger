package fdhttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/foodora/go-ranger/fdhttp"
)

type myHandler struct{}

func (h *myHandler) Init(router *fdhttp.Router) {
	router.GET("/get", h.Get)
	// router.POST("/post", h.Post)
	// router.PUT("/put", h.Put)
	// router.DELETE("/delete", h.Delete)
	router.StdGET("/std-get", h.StdGet)
	// router.StdPOST("/std-post", h.StdPost)
	// router.StdPUT("/std-put", h.StdPut)
	// router.StdDELETE("/std-delete", h.StdDelete)
}

func (h *myHandler) Get(ctx context.Context) (int, interface{}, error) {
	return http.StatusNotImplemented, nil, errors.New("method not implemented")
}

func (h *myHandler) StdGet(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	err := errors.New("method not implemented")

	resp := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(resp)
}

// ... All other method handlers

func ExampleHandler() {
	router := fdhttp.NewRouter()
	router.Register(&myHandler{})
}
