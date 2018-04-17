package fdhttp

import "net/http"

// Middleware is the method signature used as a wrapper of http request
type Middleware func(next http.Handler) http.Handler
