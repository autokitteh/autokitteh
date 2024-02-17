package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLazy1(t *testing.T) {
	n := 0

	l := Lazy1(func(int) int { n++; return n }, 0)

	assert.Equal(t, l(), 1)
	assert.Equal(t, l(), 1)
}
