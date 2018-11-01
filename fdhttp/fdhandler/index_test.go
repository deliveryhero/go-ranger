package fdhandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdhandler"
	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	indexHandler := fdhandler.NewIndex()
	indexHandler.Path = "/dir"

	router := fdhttp.NewRouter()
	router.Register(indexHandler)
	router.StdGET("/v1/foo/:id", func(w http.ResponseWriter, req *http.Request) {})
	router.PUT("/v1/foo/:id", func(ctx context.Context) (int, interface{}) {
		return http.StatusOK, nil
	})
	router.StdGET("/v1/bar/:id", func(w http.ResponseWriter, req *http.Request) {}).SetName("get_bar")
	router.PUT("/v1/bar/:id", func(ctx context.Context) (int, interface{}) {
		return http.StatusOK, nil
	}).SetName("update_bar")

	req := httptest.NewRequest(http.MethodGet, "/dir", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	assert.Equal(t, "/v1/foo/:id", resp["GET_v1_foo_id"])
	assert.Equal(t, "/v1/foo/:id", resp["PUT_v1_foo_id"])
	assert.Equal(t, "/v1/bar/:id", resp["get_bar"])
	assert.Equal(t, "/v1/bar/:id", resp["update_bar"])
}
