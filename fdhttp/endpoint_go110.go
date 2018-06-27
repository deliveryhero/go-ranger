// +build go1.10

package fdhttp

import (
	"fmt"
	"strings"
)

func (e Endpoint) PathParam(params map[string]string) string {
	var b strings.Builder

	path := strings.Split(strings.TrimPrefix(e.Path, "/"), "/")
	for _, part := range path {
		i := strings.Index(part, "*")
		if i >= 0 {
			param := params[part[i+1:]]
			fmt.Fprintf(&b, "/%s%s", part[:i], param)
			continue
		}

		i = strings.Index(part, ":")
		if i >= 0 {
			param := params[part[i+1:]]
			fmt.Fprintf(&b, "/%s%s", part[:i], param)
			continue
		}

		fmt.Fprintf(&b, "/%s", part)
	}

	return b.String()
}
