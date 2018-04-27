package fdhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func TestNewCORSMiddleware(t *testing.T) {
	origin := "https://api.foodora.com"
	corsMiddleware := fdhttp.NewCORSMiddleware(origin)

	router := fdhttp.NewRouter()
	router.Use(corsMiddleware)
	router.StdGET("/foo", func(w http.ResponseWriter, req *http.Request) {})

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewCORSMiddleware_IsIgnoredIfHandlerSetted(t *testing.T) {
	origin := "https://api.foodora.com"
	corsMiddleware := fdhttp.NewCORSMiddleware(origin)

	router := fdhttp.NewRouter()
	router.Use(corsMiddleware)
	router.GET("/foo", func(ctx context.Context) (int, interface{}) {
		fdhttp.SetResponseHeaderValue(ctx, "Access-Control-Allow-Origin", "*")
		return http.StatusOK, nil
	})

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, fdhttp.CORSOriginAll, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewCORSMiddleware_IsIgnoredIfStdHandlerSetted(t *testing.T) {
	origin := "https://api.foodora.com"
	corsMiddleware := fdhttp.NewCORSMiddleware(origin)

	router := fdhttp.NewRouter()
	router.Use(corsMiddleware)
	router.StdGET("/foo", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, fdhttp.CORSOriginAll, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewCORSHandler(t *testing.T) {
	corsHandlers := fdhttp.NewCORSHandler()
	corsHandlers.Origin = "https://api.foodora.com"
	corsHandlers.Credentials = true
	corsHandlers.ExposeHeaders = []string{
		"X-Personal-One",
		"X-Personal-Two",
	}
	corsHandlers.MaxAge = 25 * time.Minute

	router := fdhttp.NewRouter()
	router.Register(corsHandlers)
	router.StdGET("/foo", func(w http.ResponseWriter, req *http.Request) {})
	router.PUT("/foo", func(ctx context.Context) (int, interface{}) {
		return http.StatusOK, nil
	})

	req := httptest.NewRequest(http.MethodOptions, "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, corsHandlers.Origin, w.Header().Get("Access-Control-Allow-Origin"))
	assert.ElementsMatch(t, []string{
		"OPTIONS",
		"GET",
		"PUT",
	}, strings.Split(w.Header().Get("Access-Control-Allow-Methods"), ", "))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "X-Personal-One,X-Personal-Two", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "1500", w.Header().Get("Access-Control-Max-Age"))
}

func TestNewCORSHandler_WithCrdentialsDisabled(t *testing.T) {
	corsHandlers := fdhttp.NewCORSHandler()
	corsHandlers.MaxAge = 0

	router := fdhttp.NewRouter()
	router.Register(corsHandlers)

	req := httptest.NewRequest(http.MethodOptions, "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, fdhttp.CORSOriginAll, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	_, ok := w.HeaderMap["Access-Control-Allow-Credentials"]
	assert.False(t, ok)
	_, ok = w.HeaderMap["Access-Control-Allow-Headers"]
	assert.False(t, ok)
	_, ok = w.HeaderMap["Access-Control-Max-Age"]
	assert.False(t, ok)
}
