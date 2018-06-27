package fdhttp

import (
	"fmt"
	"strings"
)

// Endpoint is returned when you create a router,
// you can call Endpoint.SetName() to give a name to this
// endpoint. After that you can use this name to generate
// URL. Check Router.Path() and Router.PathParam()
type Endpoint struct {
	router *Router
	Name   string
	Method string
	Path   string
}

// SetName give a better name to the endpoint, otherwise
// will be GET_endpoint.
func (e *Endpoint) SetName(name string) {
	e.Name = name
	e.router.addEndpoint(e)
}

// addEndpoint save endpoint to the list of available endpoints.
// The name generate will be something like this:
// 		GET /v:version/people/:id/metadata
// 		become
// 		GET_v<version>_people_<id>_metadata
func (r *Router) addEndpoint(e *Endpoint) {
	// add endpoint to the main router
	if r.parent != nil {
		r.parent.addEndpoint(e)
		return
	}

	if e.Name == "" {
		var name strings.Builder
		name.WriteString(e.Method)
		name.WriteRune('_')

		for _, r := range strings.Trim(e.Path, "/") {
			if r == ':' || r == '*' {
				continue
			}
			if r == '/' {
				name.WriteRune('_')
				continue
			}

			name.WriteRune(r)
		}

		e.Name = name.String()
	}

	if _, ok := r.endpoints[e.Name]; !ok {
		r.endpoints[e.Name] = *e
	}
}

func (r *Router) Path(endpointName string) string {
	if r.parent != nil {
		return r.parent.Path(endpointName)
	}

	endpoint, ok := r.endpoints[endpointName]
	if !ok {
		panic(fmt.Sprintf("No endpoint with name %s was found", endpointName))
	}

	return endpoint.Path
}

func (r *Router) PathParam(endpointName string, params map[string]string) string {
	if r.parent != nil {
		return r.parent.PathParam(endpointName, params)
	}

	endpoint, ok := r.endpoints[endpointName]
	if !ok {
		panic(fmt.Sprintf("No endpoint with name %s was found", endpointName))
	}

	return endpoint.PathParam(params)
}
