package fdbackoff

import (
	"math"
	"time"
)

// Func ...
type Func func(attempt int) time.Duration

var Fixed = func(waitFor time.Duration) Func {
	return func(int) time.Duration {
		return waitFor
	}
}

var Linear = func(startWith time.Duration) Func {
	s := float64(startWith)
	return func(attempt int) time.Duration {
		return time.Duration(s * float64(attempt+1))
	}
}

var Exponential = func(startWith time.Duration) Func {
	s := float64(startWith)
	return func(attempt int) time.Duration {
		return time.Duration(s * math.Pow(2, float64(attempt)))
	}
}
