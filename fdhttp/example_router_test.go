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
	router.Prefix = "/v1"
	router.StdGET("/get", h.GetV1)
	// router.StdPOST("/post", h.PostV1)
	// router.StdPUT("/put", h.PutV1)
	// router.StdDELETE("/delete", h.DeleteV1)

	subRouter := router.SubRouter()
	subRouter.Prefix = "/v2"
	subRouter.GET("/get", h.GetV2)
	// subRouter.POST("/post", h.PostV2)
	// subRouter.PUT("/put", h.PutV2)
	// subRouter.DELETE("/delete", h.DeleteV2)
}

func (h *myHandler) GetV1(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	err := errors.New("method not implemented")

	resp := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *myHandler) GetV2(ctx context.Context) (int, interface{}) {
	return http.StatusNotImplemented, fdhttp.Error{
		Code:    "not_implemented",
		Message: "Method not implemented",
	}
}

// ... All other method handlers

func ExampleRouter() {
	router := fdhttp.NewRouter()
	router.Register(&myHandler{})
}
