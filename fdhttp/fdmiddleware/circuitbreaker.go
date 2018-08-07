package fdmiddleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	circuit "github.com/rubyist/circuitbreaker"
)

// CircuitWindowTime is the window time used to calculate error rate.
// It affect only circuit breaker that was not created yet.
var CircuitWindowTime = 10 * time.Second

// ErrCircuitOpen ...
var ErrCircuitOpen = circuit.ErrBreakerOpen

type Circuit struct {
	mu      sync.RWMutex
	breaker *circuit.Breaker
}

// NewCircuitBreaker receive a backoffFunc that will be used to decide
// when the circuit breaker should retry. You also need to pass the error rate\
// that will be calculate as (number of failures / total attempts). The error
// rate is calculated over a sliding window of 10 secs (by default, check DefaultWindowTime).
// Circuit will not open until there have been at least minSamples events.
func NewCircuitBreaker(backoffFunc fdbackoff.Func, rate float64, minSamples int64) *Circuit {
	breaker := circuit.NewBreakerWithOptions(&circuit.Options{
		BackOff: &circuitBreakerBackoff{
			attempt: 1,
			fn:      backoffFunc,
		},
		ShouldTrip:    nil,
		WindowTime:    CircuitWindowTime,
		WindowBuckets: circuit.DefaultWindowBuckets,
	})

	circuit := &Circuit{breaker: breaker}
	circuit.Configure(rate, minSamples)

	return circuit
}

// Configure updates the current configuration of error rates.
// rate or minSamples equal to zero, disable the circuit breaker.
func (c *Circuit) Configure(rate float64, minSamples int64) {
	var tripFunc circuit.TripFunc
	if rate > 0 && minSamples > 0 {
		tripFunc = circuit.RateTripFunc(rate, minSamples)
	}

	c.mu.Lock()
	c.breaker.ShouldTrip = tripFunc
	c.breaker.ResetCounters()
	c.mu.Unlock()
}

func (c *Circuit) Wrap(next Doer) Doer {
	return DoerFunc(func(req *http.Request) (resp *http.Response, err error) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		breakerErr := c.breaker.CallContext(req.Context(), func() error {
			resp, err = next.Do(req)
			if err != nil {
				return err
			}

			if resp != nil && resp.StatusCode >= 500 {
				return fmt.Errorf("%s %s: %s", req.Method, req.URL.String(), http.StatusText(resp.StatusCode))
			}

			return nil
		}, 0)

		if err == nil && breakerErr != nil {
			err = breakerErr
		}

		return
	})
}

type circuitBreakerBackoff struct {
	attempt int
	fn      fdbackoff.Func
}

func (b *circuitBreakerBackoff) NextBackOff() time.Duration {
	return b.fn(b.attempt)
}

func (b *circuitBreakerBackoff) Reset() {
	b.attempt = 1
}
