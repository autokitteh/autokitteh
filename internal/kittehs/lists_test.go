package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	assert.Equal(t, NewFilter(func(x int) bool { return x < 3 })([]int{1, 2, 3}), []int{1, 2})
	assert.Equal(t, NewFilter(func(x int) bool { return x < 3 })([]int{}), []int{})
	assert.Equal(t, NewFilter(func(int) bool { return false })([]int{1, 2, 3}), []int{})
	assert.Equal(t, NewFilter(func(int) bool { return true })([]int{1, 2, 3}), []int{1, 2, 3})
	assert.Nil(t, NewFilter(func(x int) bool { return x < 3 })(nil))
	assert.Nil(t, NewFilter[int](nil)(nil))
}

func TestContainedIn(t *testing.T) {
	f := ContainedIn(1, 2, 3)

	assert.True(t, f(1))
	assert.True(t, f(2))
	assert.True(t, f(3))
	assert.False(t, f(0))

	f = ContainedIn[int]()

	assert.False(t, f(0))
	assert.False(t, f(1))
}
