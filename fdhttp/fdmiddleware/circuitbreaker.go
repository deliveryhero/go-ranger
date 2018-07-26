package fdmiddleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	circuit "github.com/rubyist/circuitbreaker"
)

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

type Circuit struct {
	Breaker *circuit.Breaker
	// Use if you need update something in the Breaker
	BreakerMu sync.Mutex
}

// NewCircuitClientMiddleware receive a backoffFunc that will be used to decide
// if circuit breaker should retry. You also need to pass tripFunc that will be executed
// to decide if circuit should close or not.
// Once that the circuit is open, it'll call backoffFunc to the next attempt
// after the time will receive a half-open to retry.
func NewCircuitBreaker(backoffFunc fdbackoff.Func, tripFunc circuit.TripFunc) ClientMiddleware {
	breaker := circuit.NewBreakerWithOptions(&circuit.Options{
		BackOff: &circuitBreakerBackoff{
			attempt: 1,
			fn:      backoffFunc,
		},
		ShouldTrip: tripFunc,
	})

	return &Circuit{
		Breaker: breaker,
	}
}

func (c *Circuit) Wrap(next Doer) Doer {
	return DoerFunc(func(req *http.Request) (resp *http.Response, err error) {
		c.BreakerMu.Lock()
		defer c.BreakerMu.Unlock()

		breakerErr := c.Breaker.CallContext(req.Context(), func() error {
			resp, err = next.Do(req)
			if err != nil {
				return err
			}

			if resp.StatusCode >= 500 {
				return fmt.Errorf("server respond with %s", resp.Status)
			}

			return nil
		}, 0)

		if err == nil && breakerErr != nil {
			err = breakerErr
		}

		return
	})
}