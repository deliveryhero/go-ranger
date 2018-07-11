package fdbackoff_test

import (
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	"github.com/stretchr/testify/assert"
)

func TestFixed(t *testing.T) {
	fixedBackoff := fdbackoff.Fixed(2 * time.Second)

	assert.Equal(t, 0*time.Second, fixedBackoff(0))
	assert.Equal(t, 2*time.Second, fixedBackoff(1))
	assert.Equal(t, 2*time.Second, fixedBackoff(2))
	assert.Equal(t, 2*time.Second, fixedBackoff(3))
}

func TestConstant(t *testing.T) {
	constBackoff := fdbackoff.Constant(2 * time.Second)

	assert.Equal(t, 0*time.Second, constBackoff(0))
	assert.Equal(t, 2*time.Second, constBackoff(1))
	assert.Equal(t, 4*time.Second, constBackoff(2))
	assert.Equal(t, 6*time.Second, constBackoff(3))
}

func TestExponential(t *testing.T) {
	expBackoff := fdbackoff.Exponential(2 * time.Second)

	assert.Equal(t, 0*time.Second, expBackoff(0))
	assert.Equal(t, 2*time.Second, expBackoff(1))
	assert.Equal(t, 4*time.Second, expBackoff(2))
	assert.Equal(t, 8*time.Second, expBackoff(3))
}
