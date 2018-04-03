package fdhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func TestResponseJSON(t *testing.T) {
	w := httptest.NewRecorder()

	resp := struct {
		Success bool `json:"success"`
		Data    int  `json:"data"`
	}{
		Success: true,
		Data:    1,
	}

	fdhttp.ResponseJSON(w, http.StatusBadRequest, resp)

	contentType := w.HeaderMap["Content-Type"]
	assert.Len(t, contentType, 1)
	assert.Equal(t, "application/json; charset=utf-8", contentType[0])
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, `{"success":true,"data":1}`+"\n", w.Body.String())
}
