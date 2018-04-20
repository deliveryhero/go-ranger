package fdhttp

import (
	"bytes"
	"html/template"
	"net"
	"net/http"
	"time"
)

// LogMiddleware is a implementation of Middleware with some additional methods to
// be configured: SetLogger() and SetLoggerFunc()
type LogMiddleware Middleware

// LogRequest contain all necessary fields to be logged
type LogRequest struct {
	http.Request
	Response   *LogResponse
	RemoteAddr string
}

// LogByRequestFunc specify a function that will be called everytime that is necessary
// log something
type LogByRequestFunc func(logReq *LogRequest)

var RequestLogFormat = "{{.RemoteAddr}} [{{.Response.Elapsed}}] \"{{.Method}} {{.RequestURI}} {{.Proto}}\" {{.Response.StatusCode}} {{.Response.StatusText}} \"{{.UserAgent}}\""

// NewLogMiddleware create a log middleware
func NewLogMiddleware() LogMiddleware {
	logFn := withLogger(defaultLogger)
	return LogMiddleware(logMiddleware(logFn))
}

// logMiddleware function that return a real Middleware using logFn as function to log
func logMiddleware(logFn LogByRequestFunc) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			started := time.Now()

			lr := &LogResponse{ResponseWriter: w}
			next.ServeHTTP(lr, req)

			lr.Elapsed = time.Since(started)

			logReq := &LogRequest{
				Request:    *req,
				Response:   lr,
				RemoteAddr: getRemoteAddr(req),
			}

			logFn(logReq)
		}

		return http.HandlerFunc(fn)
	}
}

// SetLogger set a struct to log
func (m *LogMiddleware) SetLogger(log Logger) {
	logFn := withLogger(log)
	*m = LogMiddleware(logMiddleware(logFn))
}

// SetLoggerFunc set a function that is called everytime that need to log
func (m *LogMiddleware) SetLoggerFunc(fn LogByRequestFunc) {
	*m = LogMiddleware(logMiddleware(fn))
}

// Middleware return LogMiddleware instance casted to Middleware
func (m *LogMiddleware) Middleware() Middleware {
	return Middleware(*m)
}

// LogResponse it's a wrap to be able read the status code
type LogResponse struct {
	http.ResponseWriter
	StatusCode int
	Elapsed    time.Duration
}

func (lr *LogResponse) Header() http.Header {
	return lr.ResponseWriter.Header()
}

func (lr *LogResponse) Write(b []byte) (int, error) {
	return lr.ResponseWriter.Write(b)
}

func (lr *LogResponse) WriteHeader(code int) {
	lr.StatusCode = code
	lr.ResponseWriter.WriteHeader(code)
}

func (lr *LogResponse) StatusText() string {
	return http.StatusText(lr.StatusCode)
}

// withLogger is a wrap to log using fdhttp.Logger
func withLogger(log Logger) LogByRequestFunc {
	tmpl := template.Must(template.New("log-template").Parse(RequestLogFormat))

	return func(logReq *LogRequest) {
		var b bytes.Buffer
		tmpl.Execute(&b, logReq)

		log.Printf(b.String())
	}
}

func getRemoteAddr(req *http.Request) string {
	remoteAddr := req.Header.Get("X-Forwarded-For")
	if remoteAddr == "" {
		remoteAddr, _, _ = net.SplitHostPort(req.RemoteAddr)
	}

	return remoteAddr
}
