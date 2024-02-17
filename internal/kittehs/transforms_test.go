package kittehs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	strs = []string{"zero", "one", "two", "three", "four", "five"}
	ints = []int{0, 1, 2, 3, 4, 5}
)

func TestTransform(t *testing.T) {
	assert.Nil(t, Transform(nil, func(int) int { return 0 }))
	assert.Equal(t, Transform([]int{}, func(int) int { return 0 }), []int{})
	assert.Equal(t, Transform(ints, func(i int) string { return strs[i] }), strs)
	assert.Equal(t, Transform(ints[1:4], func(i int) int { return ints[i-1] }), ints[:3])
}

func TestTransformError(t *testing.T) {
	xs, err := TransformError(nil, func(int) (int, error) { return 0, nil })
	assert.NoError(t, err)
	assert.Nil(t, xs)

	xs, err = TransformError([]int{}, func(int) (int, error) { return 0, nil })
	if assert.NoError(t, err) {
		assert.Equal(t, xs, []int{})
	}

	ys, err := TransformError(ints, func(i int) (string, error) { return strs[i], nil })
	if assert.NoError(t, err) {
		assert.Equal(t, ys, strs)
	}

	xs, err = TransformError(ints[1:4], func(i int) (int, error) { return ints[i-1], nil })
	if assert.NoError(t, err) {
		assert.Equal(t, xs, ints[:3])
	}

	errHiss := errors.New("hiss")

	xs, err = TransformError(ints, func(i int) (int, error) {
		if i > 3 {
			return 0, errHiss
		}
		return i, nil
	})

	expected := fmt.Errorf("%w: %v", errHiss, 4)
	assert.Equal(t, err, expected)
	assert.Nil(t, xs)

	xs, err = TransformError(ints[:2], func(i int) (int, error) {
		if i > 3 {
			return 0, errors.New("boo!")
		}
		return i, nil
	})

	if assert.NoError(t, err) {
		assert.Equal(t, xs, ints[:2])
	}
}
