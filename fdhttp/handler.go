package fdhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// A type that satisfies fdhttp.Handler can be registered as a handler on fdhttp.Router.
type Handler interface {
	// This method will be called right before your fdhttp.Server run or when router.Init()
	// is called and it needes to register all endpoints that your handler implements.
	Init(*Router)
}

// EndpointFunc is the method signature to deal with http requests.
//
// See Also
//
// Router.GET(), Router.POST(), Router.PUT(), Router.DELETE()
// or functions that are compatible with standard library
// Router.StdGET(), Router.StdPOST(), Router.StdPUT(), Router.StdDELETE()
type EndpointFunc func(context.Context) (int, interface{})

type JSONer interface {
	JSON() interface{}
}

// contextKey is a value for use with context.WithValue.
type contextKey struct {
	name string
}

func (c contextKey) String() string {
	return "fdhttp context key " + c.name
}

var (
	// RouteParamContextKey is the key used to save route params.
	RouteParamContextKey = &contextKey{"route-params"}

	// RequestHeaderContextKey is the key used to save request header.
	RequestHeaderContextKey = &contextKey{"request-header"}

	// RequestBodyContextKey is the key used to save the body request.
	RequestBodyContextKey = &contextKey{"request-body"}

	// RequestFormContextKey is the key used to save the form with query string and post data.
	RequestFormContextKey = &contextKey{"request-form"}

	// RequestPostFormContextKey is the key used to save the post data.
	RequestPostFormContextKey = &contextKey{"request-post-form"}

	// ResponseHeaderContextKey is the key used to save the response header.
	ResponseHeaderContextKey = &contextKey{"response-header"}

	// ResponseErrorContextKey is the key used to save the response error.
	ResponseErrorContextKey = &contextKey{"response-error"}
)

// RouteParams set route params to context.
func RouteParams(ctx context.Context) map[string]string {
	v, _ := ctx.Value(RouteParamContextKey).(map[string]string)
	return v
}

// RouteParam get route params from context.
func SetRouteParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, RouteParamContextKey, params)
}

// RouteParam get a specific route param from context.
func RouteParam(ctx context.Context, param string) string {
	v, _ := RouteParams(ctx)[param]
	return v
}

// RequestHeader get header from context.
func RequestHeader(ctx context.Context) http.Header {
	header, _ := ctx.Value(RequestHeaderContextKey).(http.Header)
	if header == nil {
		return http.Header{}
	}
	return header
}

// RequestHeaderValue call header.Get for you without get the whole object using RequestHeader
func RequestHeaderValue(ctx context.Context, key string) string {
	header := RequestHeader(ctx)
	return header.Get(key)
}

// SetRequestHeader set header into context.
func SetRequestHeader(ctx context.Context, value http.Header) context.Context {
	return context.WithValue(ctx, RequestHeaderContextKey, value)
}

// RequestBody get body from context.
func RequestBody(ctx context.Context) io.Reader {
	body, _ := ctx.Value(RequestBodyContextKey).(io.Reader)
	return body
}

// RequestBodyJSON get body from context but deconding as JSON.
func RequestBodyJSON(ctx context.Context, v interface{}) error {
	body := RequestBody(ctx)
	return json.NewDecoder(body).Decode(v)
}

// SetRequestBody set body into context.
func SetRequestBody(ctx context.Context, value io.Reader) context.Context {
	return context.WithValue(ctx, RequestBodyContextKey, value)
}

// RequestForm get form from context.
func RequestForm(ctx context.Context) url.Values {
	form, _ := ctx.Value(RequestFormContextKey).(url.Values)
	if form == nil {
		return url.Values{}
	}
	return form
}

// RequestFormValue call form.Get for you without get the whole object using RequestForm
func RequestFormValue(ctx context.Context, key string) string {
	form := RequestForm(ctx)
	return form.Get(key)
}

// SetRequestForm set form into context.
func SetRequestForm(ctx context.Context, value url.Values) context.Context {
	return context.WithValue(ctx, RequestFormContextKey, value)
}

// RequestPostForm get form from context.
func RequestPostForm(ctx context.Context) url.Values {
	form, _ := ctx.Value(RequestPostFormContextKey).(url.Values)
	if form == nil {
		return url.Values{}
	}
	return form
}

// RequestPostFormValue call form.Get for you without get the whole object using RequestPostForm
func RequestPostFormValue(ctx context.Context, key string) string {
	form := RequestPostForm(ctx)
	return form.Get(key)
}

// SetRequestPostForm set form into context.
func SetRequestPostForm(ctx context.Context, value url.Values) context.Context {
	return context.WithValue(ctx, RequestPostFormContextKey, value)
}

// ResponseError get response error from context.
func ResponseError(ctx context.Context) *Error {
	respErr, _ := ctx.Value(ResponseErrorContextKey).(*Error)
	return respErr
}

// SetResponseError set response error into context.
func SetResponseError(ctx context.Context, respErr *Error) context.Context {
	return context.WithValue(ctx, ResponseErrorContextKey, respErr)
}

// ResponseHeader get header from context.
func ResponseHeader(ctx context.Context) http.Header {
	header, _ := ctx.Value(ResponseHeaderContextKey).(http.Header)
	return header
}

// SetResponseHeader set header into context.
func SetResponseHeader(ctx context.Context, value http.Header) context.Context {
	return context.WithValue(ctx, ResponseHeaderContextKey, value)
}

// SetResponseHeaderValue call header.Set for you without get the whole object using ResponseHeader
func SetResponseHeaderValue(ctx context.Context, key, value string) {
	header := ResponseHeader(ctx)
	header.Set(key, value)
}

// AddResponseHeaderValue call header.Add for you without get the whole object using ResponseHeader
func AddResponseHeaderValue(ctx context.Context, key, value string) {
	header := ResponseHeader(ctx)
	header.Add(key, value)
}
