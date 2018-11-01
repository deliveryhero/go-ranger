package fdmiddleware_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	circuitbreaker := fdmiddleware.NewCircuitBreakerTransport(
		fdbackoff.Linear(time.Millisecond),
		0.4, // 40% error rate
		5,
	)

	expectedErr := errors.New("failed connecting to remote server")

	var transport http.RoundTripper

	var called int
	transport = fdmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "http://localhost", req.URL.String())

		called++
		switch called {
		case 1:
			return &http.Response{StatusCode: http.StatusCreated}, nil
		case 2:
			return &http.Response{StatusCode: http.StatusInternalServerError}, nil
		case 3:
			return &http.Response{StatusCode: http.StatusAccepted}, nil
		case 4:
			return &http.Response{StatusCode: http.StatusBadRequest}, nil
		case 5:
			return nil, expectedErr
		}

		return &http.Response{StatusCode: http.StatusOK}, nil
	})

	transport = circuitbreaker.Wrap(transport)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = transport.RoundTrip(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	resp, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	resp, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = transport.RoundTrip(req)
	assert.Equal(t, expectedErr, err)

	resp, err = transport.RoundTrip(req)
	assert.Equal(t, fdmiddleware.ErrCircuitOpen, err)
}

func TestCircuitBreaker_CanceledContextDoesNotTripCircuit(t *testing.T) {
	circuit := fdmiddleware.NewCircuitBreakerTransport(fdbackoff.Linear(time.Millisecond), 1.0, 5)

	var (
		transport http.RoundTripper
		called    int
	)

	transport = fdmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called++
		return nil, nil
	})
	transport = circuit.Wrap(transport)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(req.Context())
	// cancel the context now.
	cancel()

	req = req.WithContext(ctx)

	for i := 0; i < 10; i++ {
		_, err := transport.RoundTrip(req)
		assert.NoError(t, err)
	}

	assert.Equal(t, 10, called)
}

func TestCircuitBreaker_CallWithCircuitOpenReturnBreakerError(t *testing.T) {
	circuit := fdmiddleware.NewCircuitBreakerTransport(fdbackoff.Linear(time.Millisecond), 1.0, 5)

	expectedErr := errors.New("failed connecting to remote server")

	var (
		transport http.RoundTripper
		called    int
	)

	transport = fdmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called++
		return nil, expectedErr
	})
	transport = circuit.Wrap(transport)

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		_, err := transport.RoundTrip(req)
		if i < 5 {
			assert.Equal(t, expectedErr, err)
			continue
		}
		assert.Equal(t, fdmiddleware.ErrCircuitOpen, err)
	}

	assert.Equal(t, 5, called)
}
