package fdhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Router keep a list of handlers and middlewares and it can be used
// as ServerMux to standard library.
type Router struct {
	// NotFoundHandler by default is NewNotFoundHandler
	NotFoundHandler http.HandlerFunc
	// MethodNotAllowedHandler by default is NewMethodNotAllowedHandler
	MethodNotAllowedHandler http.HandlerFunc
	// PanicHandler by default is NewPanicHandler
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})

	httprouter  *httprouter.Router
	middlewares []Middleware
	handlers    []Handler
	rootHandler http.Handler
}

var _ http.Handler = &Router{}

// NewRouter create a new route instance
func NewRouter() *Router {
	return &Router{
		httprouter: httprouter.New(),
	}
}

// Init call Init() from all handlers
func (r *Router) Init() {
	for _, h := range r.handlers {
		h.Init(r)
	}

	// Set default not found handlers
	if r.NotFoundHandler != nil {
		r.httprouter.NotFound = r.NotFoundHandler
	} else {
		r.httprouter.NotFound = NewNotFoundHandler()
	}
	// Set default method not allowed handler
	if r.MethodNotAllowedHandler != nil {
		r.httprouter.MethodNotAllowed = r.MethodNotAllowedHandler
	} else {
		r.httprouter.MethodNotAllowed = NewMethodNotAllowedHandler()
	}
	// Set default panic handler
	if r.PanicHandler != nil {
		r.httprouter.PanicHandler = r.PanicHandler
	} else {
		r.httprouter.PanicHandler = NewPanicHandler()
	}

	// build root handler with all middlewares
	r.rootHandler = r.httprouter
	for k := range r.middlewares {
		r.rootHandler = r.middlewares[len(r.middlewares)-1-k](r.rootHandler)
	}
}

// Use a middleware to wrap all http request
func (r *Router) Use(m Middleware) {
	r.middlewares = append(r.middlewares, m)
}

// Register a handler that need to register all its own routes
func (r *Router) Register(h Handler) {
	r.handlers = append(r.handlers, h)
}

// StdGET register a standard http.HandlerFunc to handle GET method
func (r *Router) StdGET(path string, handler http.HandlerFunc) {
	r.httprouter.Handler("GET", path, handler)
}

// StdPOST register a standard http.HandlerFunc to handle POST method
func (r *Router) StdPOST(path string, handler http.HandlerFunc) {
	r.httprouter.Handler("POST", path, handler)
}

// StdPUT register a standard http.HandlerFunc to handle PUT method
func (r *Router) StdPUT(path string, handler http.HandlerFunc) {
	r.httprouter.Handler("PUT", path, handler)
}

// StdDELETE register a standard http.HandlerFunc to handle DELETE method
func (r *Router) StdDELETE(path string, handler http.HandlerFunc) {
	r.httprouter.Handler("DELETE", path, handler)
}

// Handler register the method and path with fdhttp.EndpointFunc
func (r *Router) Handler(method, path string, fn EndpointFunc) {
	r.httprouter.Handle(method, path, func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := req.Context()

		// Inject route param on ctx
		for _, param := range params {
			ctx = SetRouteParam(ctx, param.Key, param.Value)
		}

		ctx = SetRequestHeader(ctx, req.Header)

		// Inject body on ctx
		if req.Body != nil {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				ResponseJSON(w, http.StatusBadRequest, &ResponseError{
					Code:    "invalid_body",
					Message: err.Error(),
				})
				return
			}

			ctx = SetRequestBody(ctx, bytes.NewBuffer(body))
			defer req.Body.Close()
		}

		// call user handler
		statusCode, resp, err := fn(ctx)
		if err != nil {
			// ignore response in case of error

			if respErr, ok := err.(*ResponseError); ok {
				resp = respErr
			} else {
				resp = &ResponseError{
					Message: err.Error(),
				}
			}
		}

		// Even in error case send all headers setted
		header := RequestHeader(ctx)
		for header, values := range header {
			for _, value := range values {
				w.Header().Set(header, value)
			}
		}

		ResponseJSON(w, statusCode, resp)
	})
}

// GET register a fdhttp.EndpointFunc to handle GET method
func (r *Router) GET(path string, fn EndpointFunc) {
	r.Handler("GET", path, fn)
}

// POST register a fdhttp.EndpointFunc to handle POST method
func (r *Router) POST(path string, fn EndpointFunc) {
	r.Handler("POST", path, fn)
}

// PUT register a fdhttp.EndpointFunc to handle PUT method
func (r *Router) PUT(path string, fn EndpointFunc) {
	r.Handler("PUT", path, fn)
}

// DELETE register a fdhttp.EndpointFunc to handle DELETE method
func (r *Router) DELETE(path string, fn EndpointFunc) {
	r.Handler("DELETE", path, fn)
}

// ServeHTTP makes this struct a valid implementation of http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.rootHandler == nil {
		r.Init()
	}

	r.rootHandler.ServeHTTP(w, req)
}
