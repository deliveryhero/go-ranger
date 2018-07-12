package fdmiddleware

import (
	"fmt"
	"net/http"
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
}

// NewCircuitClientMiddleware receive a backoffFunc that will be used to decide
// if circuit breaker should retry. You also need to pass tripFunc that will be executed
// to decide if circuit should close or not.
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
		c.Breaker.CallContext(req.Context(), func() error {
			resp, err = next.Do(req)
			if err != nil {
				return err
			}

			if resp.StatusCode >= 500 {
				return fmt.Errorf("server respond with %s", resp.Status)
			}

			return nil
		}, 0)

		return
	})
}
