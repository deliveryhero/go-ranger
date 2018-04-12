package fdhttp_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func TestRouteParam(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, fdhttp.RouteParamPrefixKey+"get_test", "value")
	assert.Equal(t, "value", fdhttp.RouteParam(ctx, "get_test"))
}

func TestRouteParam_Empty(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", fdhttp.RouteParam(ctx, "unknown"))
}

func TestSetRouteParam(t *testing.T) {
	ctx := context.Background()
	ctx = fdhttp.SetRouteParam(ctx, "set_test", "value")
	value, _ := ctx.Value(fdhttp.RouteParamPrefixKey + "set_test").(string)
	assert.Equal(t, "value", value)
}

func TestGetAndSetRouteParam(t *testing.T) {
	ctx := context.Background()
	ctx = fdhttp.SetRouteParam(ctx, "get_set_test", "value")
	assert.Equal(t, "value", fdhttp.RouteParam(ctx, "get_set_test"))
}

func TestRequestBody(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, fdhttp.RequestBodyKey, bytes.NewBufferString("value"))

	b := fdhttp.RequestBody(ctx).(*bytes.Buffer)
	assert.Equal(t, "value", b.String())
}

func TestRequestBody_Empty(t *testing.T) {
	ctx := context.Background()
	b := fdhttp.RequestBody(ctx)
	assert.Nil(t, b)
}

func TestSetRequestBody(t *testing.T) {
	ctx := context.Background()
	ctx = fdhttp.SetRequestBody(ctx, bytes.NewBufferString("value"))
	b, _ := ctx.Value(fdhttp.RequestBodyKey).(*bytes.Buffer)
	assert.Equal(t, "value", b.String())
}

func TestGetAndSetRequestBody(t *testing.T) {
	ctx := context.Background()
	ctx = fdhttp.SetRequestBody(ctx, bytes.NewBufferString("value"))
	b := fdhttp.RequestBody(ctx).(*bytes.Buffer)
	assert.Equal(t, "value", b.String())
}

func TestRequestBodyJSON(t *testing.T) {
	ctx := context.Background()
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

func TestRequestHeader(t *testing.T) {
	header := http.Header{}
	header.Add("X-Personal", "1")
	header.Add("X-Personal", "2")
	header.Add("Content-Type", "2")

	ctx := context.Background()
	ctx = context.WithValue(ctx, fdhttp.RequestHeaderKey, header)

	h := fdhttp.RequestHeader(ctx)
	assert.Equal(t, header, h)
}

func TestRequestHeader_Empty(t *testing.T) {
	ctx := context.Background()
	h := fdhttp.RequestHeader(ctx)
	assert.Nil(t, h)
}

func TestSetRequestHeader(t *testing.T) {
	header := http.Header{}
	header.Add("X-Personal", "1")
	header.Add("X-Personal", "2")
	header.Add("Content-Type", "2")

	ctx := context.Background()
	ctx = fdhttp.SetRequestHeader(ctx, header)
	h, _ := ctx.Value(fdhttp.RequestHeaderKey).(http.Header)
	assert.Equal(t, header, h)
}

func TestGetAndSetRequestHeader(t *testing.T) {
	header := http.Header{}
	header.Add("X-Personal", "1")
	header.Add("X-Personal", "2")
	header.Add("Content-Type", "2")

	ctx := context.Background()
	ctx = fdhttp.SetRequestHeader(ctx, header)
	h := fdhttp.RequestHeader(ctx)
	assert.Equal(t, header, h)
}

func TestResponseHeader(t *testing.T) {
	header := http.Header{}
	header.Add("X-Personal", "1")
	header.Add("X-Personal", "2")
	header.Add("Content-Type", "2")

	ctx := context.Background()
	ctx = context.WithValue(ctx, fdhttp.ResponseHeaderKey, header)

	h := fdhttp.ResponseHeader(ctx)
	assert.Equal(t, header, h)
}

func TestResponseHeader_Empty(t *testing.T) {
	ctx := context.Background()
	h := fdhttp.ResponseHeader(ctx)
	assert.Nil(t, h)
}

func TestSetResponseHeader(t *testing.T) {
	header := http.Header{}
	header.Add("X-Personal", "1")
	header.Add("X-Personal", "2")
	header.Add("Content-Type", "2")

	ctx := context.Background()
	ctx = fdhttp.SetResponseHeader(ctx, header)
	h, _ := ctx.Value(fdhttp.ResponseHeaderKey).(http.Header)
	assert.Equal(t, header, h)
}

func TestGetAndSetResponseHeader(t *testing.T) {
	header := http.Header{}
	header.Add("X-Personal", "1")
	header.Add("X-Personal", "2")
	header.Add("Content-Type", "2")

	ctx := context.Background()
	ctx = fdhttp.SetResponseHeader(ctx, header)
	h := fdhttp.ResponseHeader(ctx)
	assert.Equal(t, header, h)
}
