package sdkclients

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLazyCache(t *testing.T) {
	n := 0
	counter := func(int) int { n++; return n }

	f := lazyCache(counter, 0)

	assert.Equal(t, f(), 1)
	assert.Equal(t, f(), 1)
}
