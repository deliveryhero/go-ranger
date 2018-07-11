package fdbackoff

import (
	"math"
	"time"
)

// Func return to the caller the duration that need to
// wait between the next call.
// All implementation return 0 in case attempt == 0, so
// it makes sense start call your backoff func only after
// the first unsuccessful attempt
type Func func(attempt int) time.Duration

var Fixed = func(waitFor time.Duration) Func {
	return func(attempt int) time.Duration {
		if attempt == 0 {
			return 0
		}

		return waitFor
	}
}

var Constant = func(startWith time.Duration) Func {
	s := float64(startWith)
	return func(attempt int) time.Duration {
		return time.Duration(s * float64(attempt))
	}
}

var Exponential = func(startWith time.Duration) Func {
	s := float64(startWith)
	return func(attempt int) time.Duration {
		if attempt == 0 {
			return 0
		}

		return time.Duration(s * math.Pow(2, float64(attempt-1)))
	}
}
