package fdhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointURL_WithoutParameters(t *testing.T) {
	e := Endpoint{path: "/a/b/c"}
	assert.Equal(t, "/a/b/c", e.URL())
}

func TestEndpointURL_CatchAllParameter(t *testing.T) {
	e := Endpoint{path: "/download/ranger_*file"}
	assert.Equal(t, "/download/ranger_test.pdf", e.URLParam(map[string]string{
		"file": "test.pdf",
	}))
}

func TestEndpointURL_NamedParameter(t *testing.T) {
	e := Endpoint{path: "/v:version/:id/:state"}
	assert.Equal(t, "/v2/123/active", e.URLParam(map[string]string{
		"version": "2",
		"id":      "123",
		"state":   "active",
	}))
}
