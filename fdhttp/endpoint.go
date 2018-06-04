package fdhttp

import (
	"fmt"
)

type Endpoint struct {
	router *Router
	path   string
}

func (e Endpoint) Name(name string) {
	e.router.addNamedEndpoint(name, e)
}

func (e Endpoint) URL() string {
	return e.path
}

func (r *Router) URL(name string) string {
	if r.parent != nil {
		return r.parent.URL(name)
	}

	endpoint, ok := r.endpoints[name]
	if !ok {
		panic(fmt.Sprintf("No endpoint with name %s was found", name))
	}

	return endpoint.URL()
}

func (r *Router) URLParam(name string, params map[string]string) string {
	if r.parent != nil {
		return r.parent.URLParam(name, params)
	}

	endpoint, ok := r.endpoints[name]
	if !ok {
		panic(fmt.Sprintf("No endpoint with name %s was found", name))
	}

	return endpoint.URLParam(params)
}
