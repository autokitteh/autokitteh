package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstNonZero(t *testing.T) {
	assert.Zero(t, FirstNonZero[int]())
	assert.Zero(t, FirstNonZero[int](0, 0))
	assert.Equal(t, 1, FirstNonZero[int](1))
	assert.Equal(t, 1, FirstNonZero[int](0, 1))
	assert.Equal(t, 1, FirstNonZero[int](1, 0))
	assert.Equal(t, 2, FirstNonZero[int](0, 0, 2, 0))
}
