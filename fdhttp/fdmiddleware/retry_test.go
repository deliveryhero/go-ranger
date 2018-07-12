package fdmiddleware_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	"github.com/stretchr/testify/assert"
)

func TestRetryClientMiddleware(t *testing.T) {
	middleware := fdmiddleware.NewRetryClient(4, fdbackoff.Constant(time.Millisecond))

	c := fdhttp.NewClient()
	c.Use(middleware)

	var srvCalled int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		srvCalled++

		var status int
		switch srvCalled {
		case 1:
			status = http.StatusInternalServerError
		case 2:
			status = http.StatusTooManyRequests
		case 3:
			status = http.StatusOK
		default:
			t.Error(t, "It was not expected this call")
			status = http.StatusInternalServerError
		}

		w.WriteHeader(status)
		fmt.Fprint(w, http.StatusText(status))
	}))
	defer ts.Close()

	resp, err := c.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, 3, srvCalled)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "OK", string(body))
}
