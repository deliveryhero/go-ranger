// +build !go1.10

package fdhttp

import (
	"bytes"
	"fmt"
	"strings"
)

func (e Endpoint) buildName() string {
	var name bytes.Buffer
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

	return name.String()
}

func (e Endpoint) PathParam(params map[string]string) string {
	var b bytes.Buffer

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
