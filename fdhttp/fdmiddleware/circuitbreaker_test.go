package fdmiddleware_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	circuit "github.com/rubyist/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	circuitbreaker := fdmiddleware.NewCircuitBreaker(
		fdbackoff.Linear(time.Millisecond),
		circuit.ConsecutiveTripFunc(3),
	)

	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
	}
	expectedErr := errors.New("error")

	var doer fdmiddleware.Doer
	doer = fdmiddleware.DoerFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "http://localhost", req.URL.String())
		return expectedResp, expectedErr
	})
	doer = circuitbreaker.Wrap(doer)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	resp, err := doer.Do(req)
	assert.Equal(t, expectedResp, resp)
	assert.Equal(t, expectedErr, err)
}

func TestCircuitBreaker_ErrorCountAsFailure(t *testing.T) {
	circuit := fdmiddleware.NewCircuitBreaker(
		fdbackoff.Linear(time.Millisecond),
		circuit.ConsecutiveTripFunc(3),
	)

	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
	}
	expectedErr := errors.New("error")

	var doer fdmiddleware.Doer
	doer = fdmiddleware.DoerFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "http://localhost", req.URL.String())
		return expectedResp, expectedErr
	})
	doer = circuit.Wrap(doer)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	doer.Do(req)
	doer.Do(req)
	assert.Equal(t, int64(2), circuit.Breaker.Failures())
}

func TestCircuitBreaker_CanceledContextDoesNotCount(t *testing.T) {
	circuit := fdmiddleware.NewCircuitBreaker(
		fdbackoff.Linear(time.Millisecond),
		circuit.ConsecutiveTripFunc(3),
	)

	var doer fdmiddleware.Doer
	doer = fdmiddleware.DoerFunc(func(req *http.Request) (*http.Response, error) {
		time.Sleep(time.Second)
		return nil, errors.New("error")
	})
	doer = circuit.Wrap(doer)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(req.Context(), time.Second)
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	req = req.WithContext(ctx)

	doer.Do(req)
	assert.Equal(t, int64(0), circuit.Breaker.Failures())
}

func TestCircuitBreaker_CallWithCircuitOpenReturnBreakerError(t *testing.T) {
	circuitbreaker := fdmiddleware.NewCircuitBreaker(
		fdbackoff.Linear(time.Millisecond),
		circuit.ConsecutiveTripFunc(1),
	)

	expectedErr := errors.New("error")

	var doer fdmiddleware.Doer
	doer = fdmiddleware.DoerFunc(func(req *http.Request) (*http.Response, error) {
		time.Sleep(time.Second)
		return nil, expectedErr
	})
	doer = circuitbreaker.Wrap(doer)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	_, err = doer.Do(req)
	assert.Equal(t, expectedErr, err)

	_, err = doer.Do(req)
	assert.Equal(t, circuit.ErrBreakerOpen, err)
}
