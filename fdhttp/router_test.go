package fdhttp_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

type dummyHandler struct {
	initialized bool
	initFunc    func(r *fdhttp.Router)
}

func (eh *dummyHandler) Init(r *fdhttp.Router) {
	eh.initialized = true
	if eh.initFunc != nil {
		eh.initFunc(r)
	}
}

func newMiddleware(message string, called *bool) fdhttp.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			*called = true
			w.Write([]byte(message))
			next.ServeHTTP(w, req)
		})
	}
}

func TestRouter_HandlerAreInitialized(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{}

	r.Register(h)
	r.Init()

	assert.True(t, h.initialized)
}

func TestRouter_StdHandlerIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	var handlerCalled bool

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				handlerCalled = true
			}))
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	res.Body.Close()

	assert.True(t, handlerCalled)
}

func TestRouter_HandlerIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	var handlerCalled bool

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				handlerCalled = true
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	res.Body.Close()

	assert.True(t, handlerCalled)
}

func TestRouter_MiddlewareIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	var mCalled bool
	m := newMiddleware("middleware", &mCalled)

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("handler"))
			}))
		},
	}

	r.Use(m)
	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "middlewarehandler", string(body))
	res.Body.Close()

	assert.True(t, mCalled)
}

func TestRouter_MiddlewareIsCalledRightOrder(t *testing.T) {
	r := fdhttp.NewRouter()

	var m1Called, m2Called bool

	m1 := newMiddleware("m1", &m1Called)
	m2 := newMiddleware("m2", &m2Called)

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("handler"))
			}))
		},
	}

	r.Use(m1, m2)
	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "m1m2handler", string(body))

	res.Body.Close()

	assert.True(t, m1Called)
	assert.True(t, m2Called)
}

func TestRouter_RouteParamsAreSentInsideContext(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/:id", func(ctx context.Context) (int, interface{}) {
				id := fdhttp.RouteParam(ctx, "id")
				assert.Equal(t, "123", id)
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/123")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestRouter_HeadersAreSentInsideContext(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				token := fdhttp.RequestHeaderValue(ctx, "x-token")
				assert.Equal(t, "my-token", token)
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(nil))
	req.Header.Add("X-Token", "my-token")

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestRouter_FormAreSentInsideContext(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				query := fdhttp.RequestFormValue(ctx, "query")
				assert.Equal(t, "string", query)
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/?query=string")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestRouter_PostFormAreSentInsideContext(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.POST("/", func(ctx context.Context) (int, interface{}) {
				value := fdhttp.RequestPostFormValue(ctx, "field")
				assert.Equal(t, "from-body", value)
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	post := url.Values{}
	post.Add("field", "from-body")

	res, err := http.PostForm(ts.URL+"/?field=from+query+string", post)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestRouter_HeaderAreSentBackToClients(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				fdhttp.SetResponseHeaderValue(ctx, "X-Personal", "value")
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "value", res.Header.Get("x-personal"))
	res.Body.Close()
}

func TestRouter_SendResponseAsJSON(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				return http.StatusOK, map[string]interface{}{
					"success": true,
					"data": map[string]interface{}{
						"id": 123,
					},
				}
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, `{"data":{"id":123},"success":true}`+"\n", string(body))

	res.Body.Close()
}

func TestRouter_SendErrorAsJSON(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				return http.StatusOK, errors.New("my error")
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, `{"code":"","message":"my error"}`+"\n", string(body))

	res.Body.Close()
}

func TestRouter_SendResponseError(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				return http.StatusBadRequest, &fdhttp.Error{
					Code:    "123",
					Message: "something went wrong",
				}
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, `{"code":"123","message":"something went wrong"}`+"\n", string(body))

	res.Body.Close()
}

func TestRouter_ErrorIsAvailableInsideContext(t *testing.T) {
	handlerErr := &fdhttp.Error{
		Code:    "123",
		Message: "something went wrong",
	}

	r := fdhttp.NewRouter()
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				return http.StatusBadRequest, handlerErr
			})
		},
	}
	r.Register(h)

	var mCalled bool
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			mCalled = true
			next.ServeHTTP(w, req)
			assert.Equal(t, handlerErr, fdhttp.ResponseError(req.Context()))
		})
	})

	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.True(t, mCalled)
	res.Body.Close()
}

func TestRouter_NotFoundHandler(t *testing.T) {
	r := fdhttp.NewRouter()
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, `{"code":"not_found","message":"URL '/' was not found"}`+"\n", string(body))

	res.Body.Close()
}

func TestRouter_MethodNotAllowedHandler(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				return http.StatusBadRequest, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Post(ts.URL+"/", "", nil)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, `{"code":"method_not_allowed","message":"Method 'POST' is not allowed to access '/'"}`+"\n", string(body))

	res.Body.Close()
}

func TestRouter_PanicHandler(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				panic(errors.New("something bad happended"))
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, `{"code":"panic","message":"something bad happended"}`+"\n", string(body))

	res.Body.Close()
}
