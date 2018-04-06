package fdhttp_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

type dummyHealthCheck struct {
	HealthCheckFunc func() (interface{}, error)
}

func (c *dummyHealthCheck) HealthCheck() (interface{}, error) {
	return c.HealthCheckFunc()
}

type dummyHealthCheckError struct {
	DetailErrors []string
}

func (c *dummyHealthCheckError) Detail() interface{} {
	return c.DetailErrors
}

func (c *dummyHealthCheckError) Error() string {
	return "errors: " + strings.Join(c.DetailErrors, ", ")
}

func TestHealthCheckHandler(t *testing.T) {
	h := fdhttp.NewHealthCheckHandler("1.0.0", "c6053cf")

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhttp.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.True(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)
}

func TestHealthCheckHandle_WithPrefixAndDifferentURL(t *testing.T) {
	defaultHealthCheckURL := fdhttp.DefaultHealthCheckURL
	defer func() {
		fdhttp.DefaultHealthCheckURL = defaultHealthCheckURL
	}()
	fdhttp.DefaultHealthCheckURL = "/health-check"

	h := fdhttp.NewHealthCheckHandler("1.0.0", "c6053cf")
	h.Prefix = "/prefix"

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/prefix/health-check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhttp.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.True(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)
}

func TestHealthCheckHandle_SuccessfulCheck(t *testing.T) {
	dummyCheck := &dummyHealthCheck{
		HealthCheckFunc: func() (interface{}, error) {
			return map[string]interface{}{
				"active": 5,
				"idle":   2,
			}, nil
		},
	}

	h := fdhttp.NewHealthCheckHandler("1.0.0", "c6053cf")
	h.Register("dummy", dummyCheck)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhttp.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.True(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)
	assert.True(t, healthResp.Checks["dummy"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy"].Elapsed)

	detail := healthResp.Checks["dummy"].Detail.(map[string]interface{})
	assert.Equal(t, float64(5), detail["active"])
	assert.Equal(t, float64(2), detail["idle"])
}

func TestHealthCheckHandle_FailedCheck(t *testing.T) {
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
			return nil, &dummyHealthCheckError{
				DetailErrors: []string{"error1", "error2"},
			}
		},
	}

	h := fdhttp.NewHealthCheckHandler("1.0.0", "c6053cf")
	h.Register("dummy1", dummyCheck1)
	h.Register("dummy2", dummyCheck2)
	h.Register("dummy3", dummyCheck3)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var healthResp fdhttp.HealthCheckResponse

	json.NewDecoder(resp.Body).Decode(&healthResp)
	defer resp.Body.Close()

	assert.False(t, healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version.Tag)
	assert.Equal(t, "c6053cf", healthResp.Version.Commit)
	assert.Equal(t, time.Duration(0), healthResp.Elapsed)

	assert.True(t, healthResp.Checks["dummy1"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy1"].Elapsed)
	detail := healthResp.Checks["dummy1"].Detail.(map[string]interface{})
	assert.Equal(t, float64(5), detail["active"])
	assert.Equal(t, float64(2), detail["idle"])

	assert.False(t, healthResp.Checks["dummy2"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy2"].Elapsed)
	assert.Equal(t, "cannot access remote server", healthResp.Checks["dummy2"].Error)

	assert.False(t, healthResp.Checks["dummy3"].Status)
	assert.Equal(t, time.Duration(0), healthResp.Checks["dummy3"].Elapsed)
	errDetail := healthResp.Checks["dummy3"].Error.([]interface{})
	assert.Len(t, errDetail, 2)
	assert.Equal(t, "error1", errDetail[0])
	assert.Equal(t, "error2", errDetail[1])
}

func TestHealthCheckHandle_SuccessfulCheckSpecificService(t *testing.T) {
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

	h := fdhttp.NewHealthCheckHandler("1.0.0", "c6053cf")
	h.Register("dummy1", dummyCheck1)
	h.Register("dummy2", dummyCheck2)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check/dummy1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp fdhttp.HealthCheckResponse

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

func TestHealthCheckHandle_FailedCheckSpecificService(t *testing.T) {
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

	h := fdhttp.NewHealthCheckHandler("1.0.0", "c6053cf")
	h.Register("dummy1", dummyCheck1)
	h.Register("dummy2", dummyCheck2)

	router := fdhttp.NewRouter()
	router.Register(h)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/check/dummy2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var healthResp fdhttp.HealthCheckResponse

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
