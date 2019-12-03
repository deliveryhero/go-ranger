package fdhttp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
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

func newMiddleware(message string, called *bool) fdmiddleware.Middleware {
	return fdmiddleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			*called = true
			next.ServeHTTP(w, req)
			w.Write([]byte(message))
		})
	})
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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	resp.Body.Close()

	assert.True(t, handlerCalled)
}

func TestSubRouter_StdHandlerIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	var handlerCalled bool

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.SubRouter().StdGET("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				handlerCalled = true
			}))
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	resp.Body.Close()

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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	resp.Body.Close()

	assert.True(t, handlerCalled)
}

func TestSubRouter_HandlerIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	var handlerCalled bool

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.SubRouter().GET("/", func(ctx context.Context) (int, interface{}) {
				handlerCalled = true
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	resp.Body.Close()

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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handlermiddleware", string(body))
	resp.Body.Close()

	assert.True(t, mCalled)
}

func TestSubRouter_MiddlewareIsCalled(t *testing.T) {
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

	sr := r.SubRouter()
	sr.Prefix = "/prefix"
	sr.Use(m)
	sr.Register(h)

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handler", string(body))
	resp.Body.Close()
	assert.False(t, mCalled)

	resp, err = http.Get(ts.URL + "/prefix")
	assert.NoError(t, err)

	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handlermiddleware", string(body))
	resp.Body.Close()
	assert.True(t, mCalled)
}

func TestSubRouter_MiddlewareOfParentIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	var mCalled bool
	m := newMiddleware("middleware", &mCalled)

	r.Use(m)

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("handler"))
			}))
		},
	}

	sr := r.SubRouter()
	sr.Prefix = "/prefix"

	sr.Register(h)
	r.Register(h)

	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handlermiddleware", string(body))
	resp.Body.Close()
	assert.True(t, mCalled)
	mCalled = false

	resp, err = http.Get(ts.URL + "/prefix")
	assert.NoError(t, err)

	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handlermiddleware", string(body))
	resp.Body.Close()
	assert.True(t, mCalled)
}

type C struct{}

func (c C) Init(r *fdhttp.Router) {
	r.GET("/", func(ctx context.Context) (int, interface{}) {
		return 200, 1
	})
}

func TestSubRouter_MiddlewareOfParentSubrouterIsCalled(t *testing.T) {
	r := fdhttp.NewRouter()

	c := C{}

	var mCalled bool
	m := newMiddleware("middleware", &mCalled)

	subrouter := r.SubRouter()
	subrouter.Prefix = "/prefix1"
	subrouter.Use(m)

	sr2 := subrouter.SubRouter()
	sr2.Prefix = "/prefix2"

	sr2.Register(c)
	subrouter.Register(c)

	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	//resp, err := http.Get(ts.URL + "/")
	//assert.NoError(t, err)
	//
	//body, _ := ioutil.ReadAll(resp.Body)
	//assert.Equal(t, "handler", string(body))
	//resp.Body.Close()
	//assert.False(t, mCalled)
	//mCalled = false

	resp, err := http.Get(ts.URL + "/prefix1")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "1\nmiddleware", string(body))
	resp.Body.Close()
	assert.True(t, mCalled)
	mCalled = false

	resp, err = http.Get(ts.URL + "/prefix1/prefix2")
	assert.NoError(t, err)

	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "1\nmiddleware", string(body))
	resp.Body.Close()
	assert.True(t, mCalled)
}

func TestSubRouter_MiddlewareIsCalledWhenNotUseStdHandler(t *testing.T) {
	r := fdhttp.NewRouter()

	var mCalled bool
	m := newMiddleware("middleware", &mCalled)

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(_ context.Context) (int, interface{}) {
				return http.StatusCreated, strings.NewReader("handler")
			})
		},
	}

	sr := r.SubRouter()
	sr.Prefix = "/prefix"
	sr.Use(m)
	sr.Register(h)

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handler", string(body))
	resp.Body.Close()
	assert.False(t, mCalled)

	resp, err = http.Get(ts.URL + "/prefix")
	assert.NoError(t, err)

	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handlermiddleware", string(body))
	resp.Body.Close()
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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "handlerm2m1", string(body))

	resp.Body.Close()

	assert.True(t, m1Called)
	assert.True(t, m2Called)
}

func TestRouter_WithPrefix(t *testing.T) {
	r := fdhttp.NewRouter()
	r.Prefix = "/prefix"

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/v1", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handlerv1"))
			}))
			r.GET("/v2", func(ctx context.Context) (int, interface{}) {
				return http.StatusCreated, "handlerv2"
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/prefix/v1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/prefix/v2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
}

func TestRouterURL(t *testing.T) {
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/v1", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handlerv1"))
			})).SetName("getv1")

			r.GET("/v2", func(ctx context.Context) (int, interface{}) {
				return http.StatusCreated, "handlerv2"
			}).SetName("getv2")
		},
	}

	r := fdhttp.NewRouter()
	r.Register(h)
	r.Init()

	assert.Equal(t, "/v1", r.Path("getv1"))
	assert.Equal(t, "/v2", r.Path("getv2"))
}

func TestRouterURL_WithParams(t *testing.T) {
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/v1/:name", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handlerv1"))
			})).SetName("getv1")

			r.GET("/v2/:name", func(ctx context.Context) (int, interface{}) {
				return http.StatusCreated, "handlerv2"
			}).SetName("getv2")
		},
	}

	r := fdhttp.NewRouter()
	r.Register(h)
	r.Init()

	assert.Equal(t, "/v1/foodora", r.PathParam("getv1", map[string]string{"name": "foodora"}))
	assert.Equal(t, "/v2/foodora", r.PathParam("getv2", map[string]string{"name": "foodora"}))
}

func TestSubRouterURL(t *testing.T) {
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/v1", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handlerv1"))
			})).SetName("getv1")

			r.GET("/v2", func(ctx context.Context) (int, interface{}) {
				return http.StatusCreated, "handlerv2"
			}).SetName("getv2")
		},
	}

	r := fdhttp.NewRouter().SubRouter()
	r.Register(h)
	r.Init()

	assert.Equal(t, "/v1", r.Path("getv1"))
	assert.Equal(t, "/v2", r.Path("getv2"))
}

func TestSubRouterURL_WithParams(t *testing.T) {
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdGET("/v1/:name", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handlerv1"))
			})).SetName("getv1")

			r.GET("/v2/:name", func(ctx context.Context) (int, interface{}) {
				return http.StatusCreated, "handlerv2"
			}).SetName("getv2")
		},
	}

	r := fdhttp.NewRouter().SubRouter()
	r.Register(h)
	r.Init()

	assert.Equal(t, "/v1/foodora", r.PathParam("getv1", map[string]string{"name": "foodora"}))
	assert.Equal(t, "/v2/foodora", r.PathParam("getv2", map[string]string{"name": "foodora"}))
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

	resp, err := http.Get(ts.URL + "/123")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
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

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
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

	resp, err := http.Get(ts.URL + "/?query=string")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
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

	resp, err := http.PostForm(ts.URL+"/?field=from+query+string", post)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestRouter_BodyIsSentInsideContext(t *testing.T) {
	r := fdhttp.NewRouter()

	expectedBody := "my-body"

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.POST("/", func(ctx context.Context) (int, interface{}) {
				body := fdhttp.RequestBody(ctx)

				var buf bytes.Buffer
				buf.ReadFrom(body)

				assert.Equal(t, expectedBody, buf.String())
				return http.StatusOK, nil
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Post(ts.URL, "plain/text", bytes.NewBufferString(expectedBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestRouter_BodyCanBeReadFromRequest(t *testing.T) {
	r := fdhttp.NewRouter()

	expectedBody := "my-body"

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.StdPOST("/", func(w http.ResponseWriter, req *http.Request) {
				body := fdhttp.RequestBody(req.Context())
				assert.Equal(t, body, req.Body)

				buf, err := ioutil.ReadAll(req.Body)
				assert.NoError(t, err)

				assert.Equal(t, expectedBody, string(buf))
				w.WriteHeader(http.StatusOK)
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Post(ts.URL, "plain/text", bytes.NewBufferString(expectedBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
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

	resp, err := http.Get(ts.URL + "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "value", resp.Header.Get("x-personal"))
	resp.Body.Close()
}

func TestRouter_SendResponseAsReader(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				buf := bytes.NewBufferString("here is my response")
				return http.StatusOK, buf
			})
		},
	}

	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `here is my response`, string(body))

	resp.Body.Close()
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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"data":{"id":123},"success":true}`+"\n", string(body))

	resp.Body.Close()
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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"code":"unknown","message":"my error"}`+"\n", string(body))

	resp.Body.Close()
}

type dummyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *dummyError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *dummyError) JSON() interface{} {
	return e
}

// TestRouter_SendCustomErrorAsJSON dummyError is a valid JSONer and error,
// in this case we should return to JSONer response
func TestRouter_SendCustomErrorAsJSON(t *testing.T) {
	r := fdhttp.NewRouter()

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				return http.StatusBadRequest, &dummyError{
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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"code":"123","message":"something went wrong"}`+"\n", string(body))

	resp.Body.Close()
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

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"code":"123","message":"something went wrong"}`+"\n", string(body))

	resp.Body.Close()
}

func TestRouter_ErrorIsAvailableInsideContext(t *testing.T) {
	handlerErr := &fdhttp.Error{
		Code:    "123",
		Message: "something went wrong",
	}

	r := fdhttp.NewRouter()
	r.GET("/", func(ctx context.Context) (int, interface{}) {
		return http.StatusBadRequest, handlerErr
	})

	var mCalled bool
	r.Use(fdmiddleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			mCalled = true
			next.ServeHTTP(w, req)
			assert.Equal(t, handlerErr, fdhttp.ResponseError(req.Context()))
		})
	}))

	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.True(t, mCalled)
	resp.Body.Close()
}

func TestRouter_NotFoundHandler(t *testing.T) {
	r := fdhttp.NewRouter()
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"code":"not_found","message":"URL '/' was not found"}`+"\n", string(body))

	resp.Body.Close()
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

	resp, err := http.Post(ts.URL+"/", "", nil)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"code":"method_not_allowed","message":"Method 'POST' is not allowed to access '/'"}`+"\n", string(body))

	resp.Body.Close()
}

func TestRouter_PanicHandler(t *testing.T) {
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				panic(errors.New("something bad happended"))
			})
		},
	}

	r := fdhttp.NewRouter()
	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"code":"panic","message":"something bad happended"}`+"\n", string(body))

	resp.Body.Close()
}

func TestRouter_DoNotCancelBeforeRequestFinishes(t *testing.T) {
	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				select {
				case <-ctx.Done():
					t.Error("Request shouldn't be canceled")
				case <-time.After(1 * time.Second):
				}

				return http.StatusCreated, nil
			})
		},
	}

	r := fdhttp.NewRouter()
	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestRouter_CancelingRequest(t *testing.T) {
	var handlerCanceled bool
	canceledChan := make(chan struct{}, 1)

	h := &dummyHandler{
		initFunc: func(r *fdhttp.Router) {
			r.GET("/", func(ctx context.Context) (int, interface{}) {
				select {
				case <-ctx.Done():
					handlerCanceled = true
					canceledChan <- struct{}{}
				case <-time.After(2 * time.Second):
					t.Error("Request was canceled but router didn't notify handler")
				}

				return http.StatusOK, nil
			})
		},
	}

	r := fdhttp.NewRouter()
	r.Register(h)
	r.Init()

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	// cancel request
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	resp, err := http.DefaultClient.Do(req)
	assert.Error(t, err, context.Canceled)
	assert.Nil(t, resp)
	<-canceledChan
	assert.True(t, handlerCanceled)
}
