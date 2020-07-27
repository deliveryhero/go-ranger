package fdhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
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
	// Prefix will be added in all routes
	Prefix string

	httprouter *httprouter.Router
	parent     *Router
	childs     []*Router

	middlewares []fdmiddleware.Middleware
	handlers    []Handler
	rootHandler http.Handler

	endpoints map[string]Endpoint
}

var _ http.Handler = &Router{}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// NewRouter create a new route instance
func NewRouter() *Router {
	return &Router{
		httprouter: httprouter.New(),
		endpoints:  map[string]Endpoint{},
	}
}

// Init call Init() from all handlers
func (r *Router) Init() {
	if r.parent != nil {
		// only initialize main router
		r.parent.Init()
		return
	}

	r.initHandlers()

	// Set default not found handlers
	if r.NotFoundHandler != nil {
		r.httprouter.NotFound = http.HandlerFunc(r.NotFoundHandler)
	} else {
		r.httprouter.NotFound = newNotFoundHandler()
	}
	// Set default method not allowed handler
	if r.MethodNotAllowedHandler != nil {
		r.httprouter.MethodNotAllowed = http.HandlerFunc(r.MethodNotAllowedHandler)
	} else {
		r.httprouter.MethodNotAllowed = newMethodNotAllowedHandler()
	}
	// Set default panic handler
	if r.PanicHandler != nil {
		r.httprouter.PanicHandler = r.PanicHandler
	} else {
		r.httprouter.PanicHandler = panicHandler
	}

	// build root handler with all middlewares
	r.rootHandler = r.wrapMiddlewares(r.httprouter)
}

func (r *Router) initHandlers() {
	for _, h := range r.handlers {
		h.Init(r)
	}

	// initialize all children
	for _, sr := range r.childs {
		sr.initHandlers()
	}
}

func (r *Router) wrapMiddlewares(h http.Handler) http.Handler {
	for k := range r.middlewares {
		h = r.middlewares[len(r.middlewares)-1-k].Wrap(h)
	}

	return h
}

func (r *Router) SubRouter() *Router {
	subrouter := &Router{
		parent:     r,
		httprouter: r.httprouter,
	}
	r.childs = append(r.childs, subrouter)

	return subrouter
}

// Use a middleware to wrap all http request
func (r *Router) Use(m ...fdmiddleware.Middleware) {
	r.middlewares = append(r.middlewares, m...)
}

// Register a handler that need to register all its own routes
func (r *Router) Register(h ...Handler) {
	r.handlers = append(r.handlers, h...)
}

func convertParams(ps httprouter.Params) map[string]string {
	params := map[string]string{}
	for _, p := range ps {
		params[p.Key] = p.Value
	}
	return params
}

func injectRequestBody(ctx context.Context, req *http.Request) (context.Context, error) {
	if req.Body == nil {
		return ctx, nil
	}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return ctx, err
	}
	req.Body.Close()

	req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return SetRequestBody(ctx, req.Body), nil
}

func (r *Router) StdHandler(method, path string, handler http.HandlerFunc) *Endpoint {
	if r.parent != nil {
		// register handler to the main router and but wrap middlewares from current
		// router
		return r.parent.StdHandler(method, r.Prefix+path, r.wrapMiddlewares(handler).ServeHTTP)
	}

	r.httprouter.Handle(method, r.Prefix+path, func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		ctx := req.Context()
		ctx = SetRouteParams(ctx, convertParams(ps))

		ctx, err := injectRequestBody(ctx, req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Cannot read body request: %s", err)
			return
		}

		*req = *req.WithContext(ctx)

		handler(w, req)
		// Handler is responsible to send Header, StatusCode and Body
	})

	e := &Endpoint{
		router: r,
		Method: method,
		Path:   r.Prefix + path,
	}
	r.addEndpoint(e)

	return e
}

// StdGET register a standard http.HandlerFunc to handle GET method
func (r *Router) StdGET(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("GET", path, handler)
}

// StdPOST register a standard http.HandlerFunc to handle POST method
func (r *Router) StdPOST(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("POST", path, handler)
}

// StdPUT register a standard http.HandlerFunc to handle PUT method
func (r *Router) StdPUT(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("PUT", path, handler)
}

// StdDELETE register a standard http.HandlerFunc to handle DELETE method
func (r *Router) StdDELETE(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("DELETE", path, handler)
}

// StdOPTIONS register a standard http.HandlerFunc to handle OPTIONS method
func (r *Router) StdOPTIONS(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("OPTIONS", path, handler)
}

// StdHEAD register a standard http.HandlerFunc to handle HEAD method
func (r *Router) StdHEAD(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("HEAD", path, handler)
}

// StdPATCH register a standard http.HandlerFunc to handle PATCH method
func (r *Router) StdPATCH(path string, handler http.HandlerFunc) *Endpoint {
	return r.StdHandler("PATCH", path, handler)
}

// Handler register the method and path with fdhttp.EndpointFunc
func (r *Router) Handler(method, path string, fn EndpointFunc) *Endpoint {
	prefix := make([]string, 0)

	// build prefix
	parentRouter := r.parent
	for parentRouter != nil {
		prefix = append([]string{parentRouter.Prefix}, prefix...)
		parentRouter = parentRouter.parent
	}
	prefix = append(prefix, r.Prefix)

	r.httprouter.Handle(method, strings.Join(prefix, "")+path, func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		var handler http.Handler

		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = SetRouteParams(ctx, convertParams(ps))
			ctx = SetRequest(ctx, req)
			ctx, err := injectRequestBody(ctx, req)
			if err != nil {
				ResponseJSON(w, http.StatusBadRequest, &Error{
					Code:    "invalid_body",
					Message: err.Error(),
				})
			}

			// call user handler
			statusCode, resp := fn(ctx)
			if respErr, ok := resp.(*Error); ok {
				ctx = SetResponseError(ctx, respErr)
			} else if _, ok := resp.(JSONer); ok {
				// If resp is a JSON should have precedence to error
				// Check case test TestRouter_SendCustomErrorAsJSON
			} else if err, ok := resp.(error); ok {
				// If it's a error let's convert to fdhttp.Error and return as JSON
				respErr := &Error{
					Code:    "unknown",
					Message: err.Error(),
				}
				ctx = SetResponseError(ctx, respErr)
				resp = respErr
			}

			// Override request, with that middlewares can access ctx with
			// information added here
			*req = *req.WithContext(ctx)

			if r, ok := resp.(io.Reader); ok {
				w.WriteHeader(statusCode)
				io.Copy(w, r)
			} else {
				ResponseJSON(w, statusCode, resp)
			}
		})

		currentRouter := r
		for currentRouter.parent != nil {
			// wrap with all middlewares of parents
			handler = currentRouter.wrapMiddlewares(handler)
			currentRouter = currentRouter.parent
		}

		handler.ServeHTTP(w, req)
	})

	e := &Endpoint{
		router: r,
		Method: method,
		Path:   r.Prefix + path,
	}
	r.addEndpoint(e)

	return e
}

// GET register a fdhttp.EndpointFunc to handle GET method
func (r *Router) GET(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("GET", path, fn)
}

// POST register a fdhttp.EndpointFunc to handle POST method
func (r *Router) POST(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("POST", path, fn)
}

// PUT register a fdhttp.EndpointFunc to handle PUT method
func (r *Router) PUT(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("PUT", path, fn)
}

// DELETE register a fdhttp.EndpointFunc to handle DELETE method
func (r *Router) DELETE(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("DELETE", path, fn)
}

// OPTIONS register a fdhttp.EndpointFunc to handle OPTIONS method
func (r *Router) OPTIONS(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("OPTIONS", path, fn)
}

// HEAD register a fdhttp.EndpointFunc to handle HEAD method
func (r *Router) HEAD(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("HEAD", path, fn)
}

// PATCH register a fdhttp.EndpointFunc to handle PATCH method
func (r *Router) PATCH(path string, fn EndpointFunc) *Endpoint {
	return r.Handler("PATCH", path, fn)
}

// ServeHTTP makes this struct a valid implementation of http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.rootHandler == nil {
		r.Init()
	}

	ctx := req.Context()
	ctx = SetRequestHeader(ctx, req.Header)

	ctx = SetResponse(ctx, w)
	ctx = SetResponseHeader(ctx, w.Header())

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
}

// Endpoints return a list of all endpoints registered
func (r *Router) Endpoints() []Endpoint {
	endpoints := make([]Endpoint, 0, len(r.endpoints))
	for _, e := range r.endpoints {
		endpoints = append(endpoints, e)
	}

	return endpoints
}

// Lookup return the handle, list of params extracted from the path and
// also if you should try a trailing slash redirect
func (r *Router) Lookup(method, path string) (httprouter.Handle, map[string]string, bool) {
	h, ps, redirect := r.httprouter.Lookup(method, path)
	return h, convertParams(ps), redirect
}
