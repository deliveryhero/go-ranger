package fdhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// contextKey is a value for use with context.WithValue.
type contextKey struct {
	name string
}

func (c contextKey) String() string {
	return "fdhttp context key " + c.name
}

var (
	// RequestContextKey is the key used to the original request.
	RequestContextKey = &contextKey{"request"}

	// RouteParamContextKey is the key used to save route params.
	RouteParamContextKey = &contextKey{"route-params"}

	// RequestHeaderContextKey is the key used to save request header.
	RequestHeaderContextKey = &contextKey{"request-header"}

	// RequestBodyContextKey is the key used to save the body request.
	RequestBodyContextKey = &contextKey{"request-body"}

	// RequestFormContextKey is the key used to save the form with query string and post data.
	RequestFormContextKey = &contextKey{"request-form"}

	// RequestPostFormContextKey is the key used to save the post data. Check RequestFormContextKey.
	RequestPostFormContextKey = &contextKey{"request-post-form"}

	// ResponseContextKey is the key used to save original response.
	ResponseContextKey = &contextKey{"response"}

	// ResponseHeaderContextKey is the key used to save the response header
	// and it'll be sent to clients before request ends.
	ResponseHeaderContextKey = &contextKey{"response-header"}

	// ResponseErrorContextKey is the key used to save the response error.
	// We save this information to sent to log middleware.
	ResponseErrorContextKey = &contextKey{"response-error"}
)

// Request get http request from context.
func Request(ctx context.Context) *http.Request {
	v, _ := ctx.Value(RequestContextKey).(*http.Request)
	return v
}

// SetRequest set http request to context.
func SetRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, RequestContextKey, req)
}

// RouteParams get route params from context.
func RouteParams(ctx context.Context) map[string]string {
	v, _ := ctx.Value(RouteParamContextKey).(map[string]string)
	return v
}

// SetRouteParams set route params to context.
func SetRouteParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, RouteParamContextKey, params)
}

// RouteParam get a specific route param from context.
func RouteParam(ctx context.Context, param string) string {
	v, _ := RouteParams(ctx)[param]
	return v
}

// RequestHeader get request header from context.
func RequestHeader(ctx context.Context) http.Header {
	header, _ := ctx.Value(RequestHeaderContextKey).(http.Header)
	if header == nil {
		return http.Header{}
	}
	return header
}

// RequestHeaderValue call header.Get for you without get the whole object using fdhttp.RequestHeader
func RequestHeaderValue(ctx context.Context, key string) string {
	header := RequestHeader(ctx)
	return header.Get(key)
}

// SetRequestHeader set request header into context.
func SetRequestHeader(ctx context.Context, value http.Header) context.Context {
	return context.WithValue(ctx, RequestHeaderContextKey, value)
}

// RequestBody get request body from context.
func RequestBody(ctx context.Context) io.Reader {
	body, _ := ctx.Value(RequestBodyContextKey).(io.Reader)
	return body
}

// RequestBodyJSON get request body from context but deconding as JSON.
func RequestBodyJSON(ctx context.Context, v interface{}) error {
	body := RequestBody(ctx)
	return json.NewDecoder(body).Decode(v)
}

// SetRequestBody set request body into context.
func SetRequestBody(ctx context.Context, value io.Reader) context.Context {
	return context.WithValue(ctx, RequestBodyContextKey, value)
}

// RequestForm get request form from context.
func RequestForm(ctx context.Context) url.Values {
	form, _ := ctx.Value(RequestFormContextKey).(url.Values)
	if form == nil {
		return url.Values{}
	}
	return form
}

// RequestFormValue call form.Get for you without get the whole object using fdhttp.RequestForm
func RequestFormValue(ctx context.Context, key string) string {
	form := RequestForm(ctx)
	return form.Get(key)
}

// SetRequestForm set request form to context.
func SetRequestForm(ctx context.Context, value url.Values) context.Context {
	return context.WithValue(ctx, RequestFormContextKey, value)
}

// RequestPostForm get request form from context ignoring query strings.
func RequestPostForm(ctx context.Context) url.Values {
	form, _ := ctx.Value(RequestPostFormContextKey).(url.Values)
	if form == nil {
		return url.Values{}
	}
	return form
}

// RequestPostFormValue call form.Get for you without get the whole object using fdhttp.RequestPostForm
func RequestPostFormValue(ctx context.Context, key string) string {
	form := RequestPostForm(ctx)
	return form.Get(key)
}

// SetRequestPostForm set request form into context.
func SetRequestPostForm(ctx context.Context, value url.Values) context.Context {
	return context.WithValue(ctx, RequestPostFormContextKey, value)
}

// Response set http response from context.
func Response(ctx context.Context) http.ResponseWriter {
	v, _ := ctx.Value(ResponseContextKey).(http.ResponseWriter)
	return v
}

// SetResponse set http response to context.
func SetResponse(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, ResponseContextKey, w)
}

// ResponseError get response error from context.
func ResponseError(ctx context.Context) *Error {
	respErr, _ := ctx.Value(ResponseErrorContextKey).(*Error)
	return respErr
}

// SetResponseError set response error to context.
func SetResponseError(ctx context.Context, respErr *Error) context.Context {
	return context.WithValue(ctx, ResponseErrorContextKey, respErr)
}

// ResponseHeader get response header from context.
func ResponseHeader(ctx context.Context) http.Header {
	header, _ := ctx.Value(ResponseHeaderContextKey).(http.Header)
	return header
}

// SetResponseHeader set response header to context.
func SetResponseHeader(ctx context.Context, value http.Header) context.Context {
	return context.WithValue(ctx, ResponseHeaderContextKey, value)
}

// SetResponseHeaderValue call header.Set for you without get the whole object using fdhttp.ResponseHeader
func SetResponseHeaderValue(ctx context.Context, key, value string) {
	header := ResponseHeader(ctx)
	header.Set(key, value)
}

// AddResponseHeaderValue call header.Add for you without get the whole object using fdhttp.ResponseHeader
func AddResponseHeaderValue(ctx context.Context, key, value string) {
	header := ResponseHeader(ctx)
	header.Add(key, value)
}
