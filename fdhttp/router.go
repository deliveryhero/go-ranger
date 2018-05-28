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
	methods     map[string]struct{}
}

var _ http.Handler = &Router{}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// NewRouter create a new route instance
func NewRouter() *Router {
	return &Router{
		httprouter: httprouter.New(),
		methods:    make(map[string]struct{}),
	}
}

// allowMethod save this method as allowed to be returned in CORS header
func (r *Router) allowMethod(method string) {
	if _, ok := r.methods[method]; !ok {
		r.methods[method] = struct{}{}
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
func (r *Router) Use(m ...Middleware) {
	r.middlewares = append(r.middlewares, m...)
}

// Register a handler that need to register all its own routes
func (r *Router) Register(h ...Handler) {
	r.handlers = append(r.handlers, h...)
}

func (r *Router) StdHandler(method, path string, handler http.HandlerFunc) {
	r.allowMethod(method)
	r.httprouter.Handler(method, path, handler)
}

// StdGET register a standard http.HandlerFunc to handle GET method
func (r *Router) StdGET(path string, handler http.HandlerFunc) {
	r.StdHandler("GET", path, handler)
}

// StdPOST register a standard http.HandlerFunc to handle POST method
func (r *Router) StdPOST(path string, handler http.HandlerFunc) {
	r.StdHandler("POST", path, handler)
}

// StdPUT register a standard http.HandlerFunc to handle PUT method
func (r *Router) StdPUT(path string, handler http.HandlerFunc) {
	r.StdHandler("PUT", path, handler)
}

// StdDELETE register a standard http.HandlerFunc to handle DELETE method
func (r *Router) StdDELETE(path string, handler http.HandlerFunc) {
	r.StdHandler("DELETE", path, handler)
}

// StdOPTIONS register a standard http.HandlerFunc to handle OPTIONS method
func (r *Router) StdOPTIONS(path string, handler http.HandlerFunc) {
	r.StdHandler("OPTIONS", path, handler)
}

// Handler register the method and path with fdhttp.EndpointFunc
func (r *Router) Handler(method, path string, fn EndpointFunc) {
	r.allowMethod(method)
	r.httprouter.Handle(method, path, func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		ctx := req.Context()
		// Inject route param on ctx
		params := map[string]string{}
		for _, p := range ps {
			params[p.Key] = p.Value
		}
		ctx = SetRouteParams(ctx, params)

		// Inject body on ctx
		if req.Body != nil {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				ResponseJSON(w, http.StatusBadRequest, &Error{
					Code:    "invalid_body",
					Message: err.Error(),
				})
				return
			}

			ctx = SetRequestBody(ctx, bytes.NewBuffer(body))
		}

		// call user handler
		statusCode, resp := fn(ctx)
		if respErr, ok := resp.(*Error); ok {
			ctx = SetResponseError(ctx, respErr)
		} else if j, ok := resp.(JSONer); ok {
			resp = j.JSON()
		} else if err, ok := resp.(error); ok {
			// If it's a error let's convert to fdhttp.Error and return as JSON
			respErr := &Error{
				Code:    "unknown",
				Message: err.Error(),
			}
			ctx = SetResponseError(ctx, respErr)
			resp = respErr
		}

		// Even in error case send all headers setted
		headers := ResponseHeader(ctx)
		for h, values := range headers {
			for _, v := range values {
				w.Header().Add(h, v)
			}
		}

		// Override request, with that middlewares can access the whole ctx
		*req = *req.WithContext(ctx)

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

// OPTIONS register a fdhttp.EndpointFunc to handle OPTIONS method
func (r *Router) OPTIONS(path string, fn EndpointFunc) {
	r.Handler("OPTIONS", path, fn)
}

// ServeHTTP makes this struct a valid implementation of http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.rootHandler == nil {
		r.Init()
	}

	ctx := req.Context()
	ctx = SetRequestHeader(ctx, req.Header)
	ctx = SetResponseHeader(ctx, http.Header{})

	// Inject Form and PostForm
	if req.Form == nil {
		req.ParseMultipartForm(defaultMaxMemory)
		if req.Form != nil {
			ctx = SetRequestForm(ctx, req.Form)
		}
		if req.PostForm != nil {
			ctx = SetRequestPostForm(ctx, req.PostForm)
		}
	}

	r.rootHandler.ServeHTTP(w, req.WithContext(ctx))

	// TODO send header and body here, with this even standard handler can use
	// our helper function, like AddResponseHeader, SetResponseHeader
	// Problem is how identify that handler didn't send some data already
}
