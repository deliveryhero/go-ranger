package fdhttp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, fdhttp.Request(ctx))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8000/test", nil)
	ctx = fdhttp.SetRequest(ctx, req)
	assert.Equal(t, "/test", fdhttp.Request(ctx).URL.Path)
}

func TestRouteParams(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", fdhttp.RouteParams(ctx)["invalid"])

	ctx = fdhttp.SetRouteParams(ctx, map[string]string{"get": "value"})
	assert.Equal(t, "value", fdhttp.RouteParams(ctx)["get"])
	assert.Equal(t, "", fdhttp.RouteParams(ctx)["invalid"])
}

func TestRouteParam(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", fdhttp.RouteParam(ctx, "invalid"))

	ctx = fdhttp.SetRouteParams(ctx, map[string]string{"get": "value"})
	assert.Equal(t, "value", fdhttp.RouteParam(ctx, "get"))
	assert.Equal(t, "", fdhttp.RouteParam(ctx, "invalid"))
}

func TestRequestHeader(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, fdhttp.RequestHeader(ctx))
	assert.IsType(t, http.Header{}, fdhttp.RequestHeader(ctx))
	assert.Equal(t, "", fdhttp.RequestHeader(ctx).Get("invalid"))

	header := http.Header{}
	header.Set("Content-Type", "application/xml")

	ctx = fdhttp.SetRequestHeader(ctx, header)
	assert.Equal(t, header, fdhttp.RequestHeader(ctx))
	assert.Equal(t, "application/xml", fdhttp.RequestHeader(ctx).Get("Content-Type"))
	assert.Equal(t, "", fdhttp.RequestHeader(ctx).Get("invalid"))
}

func TestRequestHeaderValue(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", fdhttp.RequestHeaderValue(ctx, "invalid"))

	header := http.Header{}
	header.Set("Content-Type", "application/xml")

	ctx = fdhttp.SetRequestHeader(ctx, header)
	assert.Equal(t, "application/xml", fdhttp.RequestHeaderValue(ctx, "Content-Type"))
	assert.Equal(t, "", fdhttp.RequestHeaderValue(ctx, "invalid"))
}

func TestRequestBody(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, fdhttp.RequestBody(ctx))

	ctx = fdhttp.SetRequestBody(ctx, bytes.NewBufferString("value"))
	assert.Equal(t, "value", fdhttp.RequestBody(ctx).(*bytes.Buffer).String())
}

func TestRequestBodyJSON(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, fdhttp.RequestBody(ctx))

	ctx = fdhttp.SetRequestBody(ctx, bytes.NewBufferString(`{"success":true,"data":1}`))

	var resp struct {
		Success bool `json:"success"`
		Data    int  `json:"data"`
	}

	err := fdhttp.RequestBodyJSON(ctx, &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 1, resp.Data)
}

func TestRequestForm(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, fdhttp.RequestForm(ctx))
	assert.IsType(t, url.Values{}, fdhttp.RequestForm(ctx))
	assert.Equal(t, "", fdhttp.RequestForm(ctx).Get("invalid"))

	form := url.Values{}
	form.Set("field", "value")

	ctx = fdhttp.SetRequestForm(ctx, form)
	assert.Equal(t, form, fdhttp.RequestForm(ctx))
	assert.Equal(t, "value", fdhttp.RequestForm(ctx).Get("field"))
	assert.Equal(t, "", fdhttp.RequestForm(ctx).Get("invalid"))
}

func TestRequestFormValue(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", fdhttp.RequestFormValue(ctx, "invalid"))

	form := url.Values{}
	form.Set("field", "value")

	ctx = fdhttp.SetRequestForm(ctx, form)
	assert.Equal(t, "value", fdhttp.RequestFormValue(ctx, "field"))
	assert.Equal(t, "", fdhttp.RequestFormValue(ctx, "invalid"))
}

func TestRequestPostForm(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, fdhttp.RequestPostForm(ctx))
	assert.IsType(t, url.Values{}, fdhttp.RequestPostForm(ctx))
	assert.Equal(t, "", fdhttp.RequestPostForm(ctx).Get("invalid"))

	form := url.Values{}
	form.Set("field", "value")

	ctx = fdhttp.SetRequestPostForm(ctx, form)
	assert.Equal(t, form, fdhttp.RequestPostForm(ctx))
	assert.Equal(t, "value", fdhttp.RequestPostForm(ctx).Get("field"))
	assert.Equal(t, "", fdhttp.RequestPostForm(ctx).Get("invalid"))
}

func TestRequestPostFormValue(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", fdhttp.RequestPostFormValue(ctx, "invalid"))

	form := url.Values{}
	form.Set("field", "value")

	ctx = fdhttp.SetRequestPostForm(ctx, form)
	assert.Equal(t, "value", fdhttp.RequestPostFormValue(ctx, "field"))
	assert.Equal(t, "", fdhttp.RequestPostFormValue(ctx, "invalid"))
}

func TestResponseError(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, fdhttp.ResponseError(ctx))

	expErr := &fdhttp.Error{
		Code:    "code",
		Message: "message",
		Detail:  "detail",
	}
	ctx = fdhttp.SetResponseError(ctx, expErr)

	err := fdhttp.ResponseError(ctx)
	assert.Equal(t, expErr, err)
}

func TestResponseHeader(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, fdhttp.ResponseHeader(ctx))
	assert.IsType(t, http.Header{}, fdhttp.ResponseHeader(ctx))
	assert.Equal(t, "", fdhttp.ResponseHeader(ctx).Get("invalid"))

	header := http.Header{}
	header.Set("Content-Type", "application/xml")

	ctx = fdhttp.SetResponseHeader(ctx, header)
	assert.Equal(t, header, fdhttp.ResponseHeader(ctx))
	assert.Equal(t, "application/xml", fdhttp.ResponseHeader(ctx).Get("Content-Type"))
	assert.Equal(t, "", fdhttp.ResponseHeader(ctx).Get("invalid"))
}

func TestSetResponseHeaderValue(t *testing.T) {
	ctx := context.Background()

	expHeader := http.Header{}
	expHeader.Set("Content-Type", "application/xml")

	ctx = fdhttp.SetResponseHeader(ctx, expHeader)
	fdhttp.SetResponseHeaderValue(ctx, "Etag", "c561c68d0ba92bbeb8b0f612a9199f722e3a621a")

	header := fdhttp.ResponseHeader(ctx)
	assert.Equal(t, "application/xml", header.Get("Content-Type"))
	assert.Equal(t, "c561c68d0ba92bbeb8b0f612a9199f722e3a621a", header.Get("Etag"))
}

func TestAddResponseHeaderValue(t *testing.T) {
	ctx := context.Background()

	expHeader := http.Header{}
	expHeader.Set("X-Personal", "1")

	ctx = fdhttp.SetResponseHeader(ctx, expHeader)
	fdhttp.AddResponseHeaderValue(ctx, "X-Personal", "2")
	fdhttp.AddResponseHeaderValue(ctx, "X-Personal", "3")

	header := fdhttp.ResponseHeader(ctx)
	assert.Len(t, header["X-Personal"], 3)
	assert.Equal(t, "1", header["X-Personal"][0])
	assert.Equal(t, "2", header["X-Personal"][1])
	assert.Equal(t, "3", header["X-Personal"][2])
}
