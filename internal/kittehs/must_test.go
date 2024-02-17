package kittehs

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust0(t *testing.T) {
	Must0(nil)

	assert.Panics(t, func() { Must0(errors.New("hiss")) })
}

func TestMust1(t *testing.T) {
	assert.Equal(t, Must1[int](1, nil), 1)

	assert.Panics(t, func() { Must1[int](1, errors.New("hiss")) })
}

func TestShoulds(t *testing.T) {
	assert.Equal(t, Should1[int](42)(0, errors.New("hiss")), 42)
	assert.Equal(t, Should1[int](42)(1, nil), 1)
	assert.Equal(t, ShouldFunc1(func(err error) int { return Must1(strconv.Atoi(err.Error())) })(1, errors.New("42")), 42)
	assert.Equal(t, ShouldFunc1(func(error) int { return 42 })(1, nil), 1)
}
