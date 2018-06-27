package fdhandler

import (
	"context"
	"net/http"
	"strings"

	"github.com/foodora/go-ranger/fdhttp"
)

// Index make available in your root page / all registered paths
// usefull if you want to expose to mobile apps for example
// routes by name.
type Index struct {
	router *fdhttp.Router
	Path   string
}

// NewIndex returns to clients a directory of all registred routes
func NewIndex() *Index {
	return &Index{}
}

func (h *Index) Init(router *fdhttp.Router) {
	h.router = router
	router.GET(h.Path, h.Index)
}

func (h *Index) Index(ctx context.Context) (int, interface{}) {
	endpoints := h.router.Endpoints()
	paths := map[string]string{}

	for _, e := range endpoints {
		if e.Method == http.MethodGet && strings.EqualFold(h.Path, e.Path) {
			continue
		}

		paths[e.Name] = e.Path
	}

	return http.StatusOK, paths
}
