package fdhttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSOriginAll can be passed to NewCORSMiddleware to accept all domains
const CORSOriginAll = "*"

type CORSHandler struct {
	router *Router

	// Origin that we accept, but defailt is setted with CORSOriginAll
	Origin string
	// Credentials control if we'll send Access-Control-Allow-Credentials or not
	Credentials bool
	// ExposeHeaders is the list of header that we can expose
	ExposeHeaders []string
	// MaxAge is setted 1 hour by default
	MaxAge time.Duration
}

func NewCORSHandler() *CORSHandler {
	return &CORSHandler{
		Origin: CORSOriginAll,
		MaxAge: 1 * time.Hour,
	}
}

func (h *CORSHandler) Init(router *Router) {
	h.router = router
	router.OPTIONS("/*anything", h.PreFlight)
	router.Use(NewCORSMiddleware(h.Origin))
}

func (h *CORSHandler) PreFlight(ctx context.Context) (int, interface{}) {
	SetResponseHeaderValue(ctx, "Access-Control-Allow-Origin", h.Origin)
	methods := make([]string, 0, len(h.router.methods))
	for k := range h.router.methods {
		methods = append(methods, k)
	}
	SetResponseHeaderValue(ctx, "Access-Control-Allow-Methods", strings.Join(methods, ", "))

	if h.Credentials {
		SetResponseHeaderValue(ctx, "Access-Control-Allow-Credentials", "true")
	}
	if len(h.ExposeHeaders) > 0 {
		SetResponseHeaderValue(ctx, "Access-Control-Allow-Headers", strings.Join(h.ExposeHeaders, ","))
	}
	if h.MaxAge > 0 {
		SetResponseHeaderValue(ctx, "Access-Control-Max-Age", strconv.FormatInt(int64(h.MaxAge/time.Second), 10))
	}

	return http.StatusOK, nil
}

// NewCORSMiddleware create a cors middleware
func NewCORSMiddleware(origin string) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			// fmt.Println(req.Method, w.Header())
			next.ServeHTTP(w, req)
			if req.Method != http.MethodOptions && w.Header().Get("Access-Control-Allow-Origin") == "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		return http.HandlerFunc(fn)
	}
}
