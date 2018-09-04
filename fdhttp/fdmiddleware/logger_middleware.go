package fdmiddleware

import (
	"bytes"
	"html/template"
	"net"
	"net/http"
	"time"
)

// Logger is the interface used internally to log
type Logger interface {
	Printf(format string, v ...interface{})
}

// RequestLogFormat is the default template used by the logger
var RequestLogFormat = "{{.RemoteAddr}} [{{.Response.Elapsed}}] \"{{.Method}} {{.RequestURI}} {{.Proto}}\" {{.Response.StatusCode}} {{.Response.StatusText}} \"{{.UserAgent}}\""

// LogByRequestFunc specify a function that will be called everytime that is necessary
// log something
type LogByRequestFunc func(logReq *LogRequest)

// LogMiddleware is a implementation of Middleware with some additional methods to
// be configured: SetLogger() and SetLoggerFunc()
type LogMiddleware struct {
	fn LogByRequestFunc
}

// NewLogMiddleware create a log middleware
func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

// SetLogger set a fdhttp.Logger to send logs
func (m *LogMiddleware) SetLogger(log Logger) {
	tmpl := template.Must(template.New("log-template").Parse(RequestLogFormat))

	m.fn = func(logReq *LogRequest) {
		var b bytes.Buffer
		tmpl.Execute(&b, logReq)
		log.Printf(b.String())
	}
}

// SetLoggerFunc set a function that is called everytime that need to log
func (m *LogMiddleware) SetLoggerFunc(fn LogByRequestFunc) {
	m.fn = fn
}

// Wrap will be called in every request
func (m *LogMiddleware) Wrap(next http.Handler) http.Handler {
	if m.fn == nil {
		panic("Using LogMiddleware without set a log function (See: SetLogger or SetLoggerFunc)")
	}

	fn := func(w http.ResponseWriter, req *http.Request) {
		started := time.Now()

		lr := &LogResponse{
			ResponseWriter: w,
			req:            req,
		}
		next.ServeHTTP(lr, req)

		lr.Elapsed = time.Since(started)

		logReq := &LogRequest{
			Request:    *req,
			Response:   lr,
			RemoteAddr: getRemoteAddr(req),
		}

		m.fn(logReq)
	}

	return http.HandlerFunc(fn)
}

// LogRequest contain all necessary fields to be logged
type LogRequest struct {
	http.Request
	Response   *LogResponse
	RemoteAddr string
}

// LogResponse it's a wrap to be able read the status code
type LogResponse struct {
	http.ResponseWriter
	req        *http.Request
	StatusCode int
	Elapsed    time.Duration
}

func (lr *LogResponse) WriteHeader(code int) {
	lr.StatusCode = code
	lr.ResponseWriter.WriteHeader(code)
}

func (lr *LogResponse) StatusText() string {
	return http.StatusText(lr.StatusCode)
}

func getRemoteAddr(req *http.Request) string {
	remoteAddr := req.Header.Get("X-Forwarded-For")
	if remoteAddr == "" {
		remoteAddr, _, _ = net.SplitHostPort(req.RemoteAddr)
	}

	return remoteAddr
}
