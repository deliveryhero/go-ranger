package fdhttp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

type dummyLog struct {
	PrintfMsg string
}

func (l *dummyLog) Printf(format string, v ...interface{}) {
	l.PrintfMsg += fmt.Sprintf(format, v...)
}

func TestNewLogMiddleware(t *testing.T) {
	logger := &dummyLog{}
	logMiddleware := fdhttp.NewLogMiddleware()
	logMiddleware.SetLogger(logger)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusBadRequest)
	}

	ts := httptest.NewServer(logMiddleware(http.HandlerFunc(handler)))
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	assert.Regexp(t, "^127.0.0.1 \\[([0-9]+\\.)?[0-9]+[nµm]?s\\] \"GET /foo HTTP/1.1\" 400 Bad Request \"Go-http-client/1.1\"$", logger.PrintfMsg)
}

func TestNewLogMiddleware_DifferentLogFormat(t *testing.T) {
	logger := &dummyLog{}
	fdhttp.RequestLogFormat = "{{.Method}} {{.RequestURI}} {{.Response.StatusCode}} {{.Response.StatusText}}"

	logMiddleware := fdhttp.NewLogMiddleware()
	logMiddleware.SetLogger(logger)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusBadRequest)
	}

	ts := httptest.NewServer(logMiddleware(http.HandlerFunc(handler)))
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	assert.Equal(t, "GET /foo 400 Bad Request", logger.PrintfMsg)
}

func TestNewLogMiddleware_CallFuncInEachRequest(t *testing.T) {
	logger := &dummyLog{}

	logMiddleware := fdhttp.NewLogMiddleware()
	logMiddleware.SetLoggerFunc(func(logReq *fdhttp.LogRequest) {
		logger.Printf("%s %s %d", logReq.Method, logReq.RequestURI, logReq.Response.StatusCode)
	})

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(logMiddleware(http.HandlerFunc(handler)))
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	assert.Equal(t, "GET /foo 200", logger.PrintfMsg)
}
