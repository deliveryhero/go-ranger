package fdhttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointPath_CatchAllParameter(t *testing.T) {
	e := Endpoint{Path: "/download/ranger_*file"}
	assert.Equal(t, "/download/ranger_test.pdf", e.PathParam(map[string]string{
		"file": "test.pdf",
	}))
}

func TestEndpointPath_NamedParameter(t *testing.T) {
	e := Endpoint{Path: "/v:version/people/:id/:state"}
	assert.Equal(t, "/v2/people/123/active", e.PathParam(map[string]string{
		"version": "2",
		"id":      "123",
		"state":   "active",
	}))
}

func TestRouteAddEndpoint(t *testing.T) {
	r := NewRouter()

	e1 := Endpoint{Method: http.MethodPost, Path: "/v1/people"}
	e2 := Endpoint{Method: http.MethodPut, Path: "/v1/people/:id"}
	e3 := Endpoint{Method: http.MethodPost, Path: "/v:version/people"}
	e4 := Endpoint{Method: http.MethodGet, Path: "/download/*file"}
	e5 := Endpoint{Method: http.MethodPost, Path: "/people/:id/activate"}

	r.addEndpoint(&e1)
	r.addEndpoint(&e2)
	r.addEndpoint(&e3)
	r.addEndpoint(&e4)
	r.addEndpoint(&e5)

	assert.Equal(t, e1, r.endpoints["POST_v1_people"])
	assert.Equal(t, e2, r.endpoints["PUT_v1_people_id"])
	assert.Equal(t, e3, r.endpoints["POST_vversion_people"])
	assert.Equal(t, e4, r.endpoints["GET_download_file"])
	assert.Equal(t, e5, r.endpoints["POST_people_id_activate"])
}
