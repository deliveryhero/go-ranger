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

// RouteParamPrefixKey is used to avoid name clashing inside of context.Context.
var RouteParamPrefixKey = "fdhttp_router_param_"

// RouteParam get router param from context.
func RouteParam(ctx context.Context, param string) string {
	v, _ := ctx.Value(RouteParamPrefixKey + param).(string)
	return v
}

// SetRouteParam set router param into context.
func SetRouteParam(ctx context.Context, param string, value interface{}) context.Context {
	return context.WithValue(ctx, RouteParamPrefixKey+param, value)
}

// RequestBodyKey is a key used inside of context.Context to save the request body
var RequestBodyKey = "fdhttp_request_body"

// RequestBody get body from context.
func RequestBody(ctx context.Context) io.Reader {
	body, _ := ctx.Value(RequestBodyKey).(io.Reader)
	return body
}

// RequestBodyJSON get body from context but deconding as JSON.
func RequestBodyJSON(ctx context.Context, v interface{}) error {
	body := RequestBody(ctx)
	return json.NewDecoder(body).Decode(v)
}

// SetRequestBody set body into context.
func SetRequestBody(ctx context.Context, value io.Reader) context.Context {
	return context.WithValue(ctx, RequestBodyKey, value)
}

// ResponseErrorKey is a key used inside of context.Context to save the request body
var ResponseErrorKey = "fdhttp_response_error"

// ResponseError get response error from context.
func ResponseError(ctx context.Context) *Error {
	respErr, _ := ctx.Value(ResponseErrorKey).(*Error)
	return respErr
}

// SetResponseError set response error into context.
func SetResponseError(ctx context.Context, respErr *Error) context.Context {
	return context.WithValue(ctx, ResponseErrorKey, respErr)
}

// RequestHeaderKey is a key used inside of context.Context to save the request headers
var RequestHeaderKey = "fdhttp_request_header"

// RequestHeader get header from context.
func RequestHeader(ctx context.Context) http.Header {
	header, _ := ctx.Value(RequestHeaderKey).(http.Header)
	return header
}

// RequestHeaderValue call header.Get for you without get the whole object using RequestHeader
func RequestHeaderValue(ctx context.Context, key string) string {
	header := RequestHeader(ctx)
	return header.Get(key)
}

// SetRequestHeader set header into context.
func SetRequestHeader(ctx context.Context, value http.Header) context.Context {
	return context.WithValue(ctx, RequestHeaderKey, value)
}

// RequestFormKey is a key used inside of context.Context to save the Request.Form
var RequestFormKey = "fdhttp_request_form"

// RequestForm get form from context.
func RequestForm(ctx context.Context) url.Values {
	form, _ := ctx.Value(RequestFormKey).(url.Values)
	return form
}

// RequestFormValue call form.Get for you without get the whole object using RequestForm
func RequestFormValue(ctx context.Context, key string) string {
	form := RequestForm(ctx)
	return form.Get(key)
}

// SetRequestForm set form into context.
func SetRequestForm(ctx context.Context, value url.Values) context.Context {
	return context.WithValue(ctx, RequestFormKey, value)
}

// RequestPostFormKey is a key used inside of context.Context to save the Request.Form
var RequestPostFormKey = "fdhttp_request_post_form"

// RequestPostForm get form from context.
func RequestPostForm(ctx context.Context) url.Values {
	form, _ := ctx.Value(RequestPostFormKey).(url.Values)
	return form
}

// RequestPostFormValue call form.Get for you without get the whole object using RequestPostForm
func RequestPostFormValue(ctx context.Context, key string) string {
	form := RequestPostForm(ctx)
	return form.Get(key)
}

// SetRequestPostForm set form into context.
func SetRequestPostForm(ctx context.Context, value url.Values) context.Context {
	return context.WithValue(ctx, RequestPostFormKey, value)
}

// ResponseHeaderKey is a key used inside of context.Context to save the request headers
var ResponseHeaderKey = "fdhttp_response_header"

// ResponseHeader get header from context.
func ResponseHeader(ctx context.Context) http.Header {
	header, _ := ctx.Value(ResponseHeaderKey).(http.Header)
	return header
}

// SetResponseHeader set header into context.
func SetResponseHeader(ctx context.Context, value http.Header) context.Context {
	return context.WithValue(ctx, ResponseHeaderKey, value)
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
