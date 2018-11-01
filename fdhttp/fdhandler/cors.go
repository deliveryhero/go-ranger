package fdhandler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
)

// CORSOriginAll can be passed to NewCORSMiddleware to accept all domains
const CORSOriginAll = "*"

type CORS struct {
	router *fdhttp.Router

	// Origin that we accept, but defailt is setted with CORSOriginAll
	Origin string
	// Credentials control if we'll send Access-Control-Allow-Credentials or not
	Credentials bool
	// Methods is the list of methods that can be accept
	Methods []string
	// ExposeHeaders is the list of header that we can expose
	ExposeHeaders []string
	// MaxAge is setted 1 hour by default
	MaxAge time.Duration
}

func NewCORS() *CORS {
	return &CORS{
		Origin: CORSOriginAll,
		MaxAge: 1 * time.Hour,
	}
}

func (h *CORS) Init(router *fdhttp.Router) {
	h.router = router
	router.OPTIONS("/*anything", h.PreFlight)
	router.Use(NewCORSMiddleware(h.Origin))
}

func (h *CORS) PreFlight(ctx context.Context) (int, interface{}) {
	fdhttp.SetResponseHeaderValue(ctx, "Access-Control-Allow-Origin", h.Origin)

	if len(h.Methods) == 0 {
		endpoints := h.router.Endpoints()
		methodsMap := make(map[string]struct{})

		for _, e := range endpoints {
			if _, ok := methodsMap[e.Method]; !ok {
				methodsMap[e.Method] = struct{}{}
				h.Methods = append(h.Methods, e.Method)
			}
		}
	}

	fdhttp.SetResponseHeaderValue(ctx, "Access-Control-Allow-Methods", strings.Join(h.Methods, ", "))

	if h.Credentials {
		fdhttp.SetResponseHeaderValue(ctx, "Access-Control-Allow-Credentials", "true")
	}
	if len(h.ExposeHeaders) > 0 {
		fdhttp.SetResponseHeaderValue(ctx, "Access-Control-Allow-Headers", strings.Join(h.ExposeHeaders, ","))
	}
	if h.MaxAge > 0 {
		fdhttp.SetResponseHeaderValue(ctx, "Access-Control-Max-Age", strconv.FormatInt(int64(h.MaxAge/time.Second), 10))
	}

	return http.StatusOK, nil
}

// NewCORSMiddleware create a cors middleware
func NewCORSMiddleware(origin string) fdmiddleware.Middleware {
	return fdmiddleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	})
}
