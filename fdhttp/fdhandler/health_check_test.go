package fdhandler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdhandler"
	"github.com/stretchr/testify/assert"
)

type dummyHealthCheck struct {
	HealthCheckFunc func() (interface{}, error)
}

func (c *dummyHealthCheck) HealthCheck(context.Context) (interface{}, error) {
	return c.HealthCheckFunc()
}

func TestHealthCheck(t *testing.T) {
	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf")

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	hostname, _ := os.Hostname()

	assert.True(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, hostname, healthResp.Hostname)
	assert.Equal(t, runtime.Version(), healthResp.System.Version)
	assert.Equal(t, runtime.NumCPU(), healthResp.System.NumCPU)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)
}

func TestHealthCheckNoSystemVersion(t *testing.T) {
	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf").DisableSystemVersion()

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse
	defer resp.Body.Close()

	assert.Empty(t, healthResp.System.Version)
}

func TestHealthCheck_WithPrefixAndDifferentURL(t *testing.T) {
	defaultHealthCheckURL := fdhandler.HealthCheckURL
	defer func() {
		fdhandler.HealthCheckURL = defaultHealthCheckURL
	}()
	fdhandler.HealthCheckURL = "/health-check"

	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf")
	h.Prefix = "/prefix"

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/prefix/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/prefix/health-check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.True(t, healthResp.Status)
}

func TestHealthCheck_SuccessfulCheck(t *testing.T) {
	dummyCheck := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return map[string]interface{}{
				"active": 5,
				"idle":   2,
			}, nil
		},
	}

	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf")
	h.Register("dummy", dummyCheck)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.True(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.True(t, healthResp.Checks["dummy"].Status)

	detail := healthResp.Checks["dummy"].Detail.(map[string]interface{})
	assert.Equal(t, float64(5), detail["active"])
	assert.Equal(t, float64(2), detail["idle"])
}

func TestHealthCheck_FailedCheck(t *testing.T) {
	fdhandler.HealthCheckServiceTimeout = 500 * time.Millisecond

	dummyCheck1 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return map[string]interface{}{
				"active": 5,
				"idle":   2,
			}, nil
		},
	}
	dummyCheck2 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return nil, errors.New("cannot access remote server")
		},
	}
	dummyCheck3 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			detail := []string{"error1", "error2"}
			return detail, errors.New("because erro1 and erro2 happend")
		},
	}
	dummyCheck4 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			time.Sleep(time.Second)
			return nil, nil
		},
	}

	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf")
	h.Register("dummy1", dummyCheck1)
	h.Register("dummy2", dummyCheck2)
	h.Register("dummy3", dummyCheck3)
	h.Register("dummy4", dummyCheck4)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.False(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)

	assert.True(t, healthResp.Checks["dummy1"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy1"].Elapsed)
	detailMap := healthResp.Checks["dummy1"].Detail.(map[string]interface{})
	assert.Equal(t, float64(5), detailMap["active"])
	assert.Equal(t, float64(2), detailMap["idle"])

	assert.False(t, healthResp.Checks["dummy2"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy2"].Elapsed)
	assert.Equal(t, "cannot access remote server", healthResp.Checks["dummy2"].Error)

	assert.False(t, healthResp.Checks["dummy3"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy3"].Elapsed)
	assert.Equal(t, "because erro1 and erro2 happend", healthResp.Checks["dummy3"].Error)
	detailArray := healthResp.Checks["dummy3"].Detail.([]interface{})
	assert.Len(t, detailArray, 2)
	assert.Equal(t, "error1", detailArray[0])
	assert.Equal(t, "error2", detailArray[1])

	assert.False(t, healthResp.Checks["dummy4"].Status)
	assert.Equal(t, "context deadline exceeded", healthResp.Checks["dummy4"].Error)
}

func TestHealthCheck_SuccessfulCheckSpecificService(t *testing.T) {
	dummyCheck1 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return map[string]interface{}{
				"active": 5,
				"idle":   2,
			}, nil
		},
	}
	dummyCheck2 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return nil, errors.New("cannot access remote server")
		},
	}

	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf")
	h.Register("dummy1", dummyCheck1)
	h.Register("dummy2", dummyCheck2)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check/dummy1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.True(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)
	assert.True(t, healthResp.Checks["dummy1"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy1"].Elapsed)

	detail := healthResp.Checks["dummy1"].Detail.(map[string]interface{})
	assert.Equal(t, float64(5), detail["active"])
	assert.Equal(t, float64(2), detail["idle"])
}

func TestHealthCheck_FailedCheckSpecificService(t *testing.T) {
	dummyCheck1 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return map[string]interface{}{
				"active": 5,
				"idle":   2,
			}, nil
		},
	}
	dummyCheck2 := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return nil, errors.New("cannot access remote server")
		},
	}

	h := fdhandler.NewHealthCheck("1.0.0", "c6053cf")
	h.Register("dummy1", dummyCheck1)
	h.Register("dummy2", dummyCheck2)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check/dummy2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var healthResp fdhandler.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.False(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)
	assert.False(t, healthResp.Checks["dummy2"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy2"].Elapsed)
	assert.Equal(t, "cannot access remote server", healthResp.Checks["dummy2"].Error)
}

func TestHealthCheck_WithoutExtra_DoesNotShowExtra(t *testing.T) {
	h := fdhandler.NewHealthCheck("1", "2")
	ctx := context.TODO()
	ctx = fdhttp.SetResponseHeader(ctx, http.Header(map[string][]string{}))

	code, response := h.Get(ctx)

	assert.Equal(t, http.StatusOK, code)
	hcResponse, ok := response.(*fdhandler.HealthCheckResponse)
	assert.True(t, ok)
	assert.Equal(t, 0, len(hcResponse.Extra))
}

func TestHealthCheck_WithExtra_ShowExtra(t *testing.T) {
	extra := map[string]string{"param1": "val1", "param2": "val2"}
	h := fdhandler.NewHealthCheck("1", "2").WithExtraParams(extra)
	ctx := context.TODO()
	ctx = fdhttp.SetResponseHeader(ctx, http.Header(map[string][]string{}))

	code, response := h.Get(ctx)

	assert.Equal(t, http.StatusOK, code)
	hcResponse, ok := response.(*fdhandler.HealthCheckResponse)
	assert.True(t, ok)
	assert.Equal(t, len(extra), len(hcResponse.Extra))
	for k, v := range extra {
		assert.Equal(t, v, hcResponse.Extra[k], fmt.Sprintf("for %s expected %s, obtained %s", k, v, hcResponse.Extra[k]))
	}
}
