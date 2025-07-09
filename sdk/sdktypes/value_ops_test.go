package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddValues(t *testing.T) {
	t.Run("integers", func(t *testing.T) {
		v1 := NewIntegerValue(2)
		v2 := NewIntegerValue(3)
		res, err := AddValues(v1, v2)
		assert.NoError(t, err)
		assert.Equal(t, NewIntegerValue(5), res)
	})

	t.Run("floats", func(t *testing.T) {
		v1 := NewFloatValue(2.5)
		v2 := NewFloatValue(3.5)
		res, err := AddValues(v1, v2)
		assert.NoError(t, err)
		assert.Equal(t, NewFloatValue(6.0), res)
	})

	t.Run("int+float", func(t *testing.T) {
		v1 := NewIntegerValue(2)
		v2 := NewFloatValue(3.5)
		res, err := AddValues(v1, v2)
		assert.NoError(t, err)
		assert.Equal(t, NewIntegerValue(5), res)
	})

	t.Run("float+int", func(t *testing.T) {
		v1 := NewFloatValue(3.5)
		v2 := NewIntegerValue(2)
		res, err := AddValues(v1, v2)
		assert.NoError(t, err)
		assert.Equal(t, NewFloatValue(5.5), res)
	})

	t.Run("strings", func(t *testing.T) {
		v1 := NewStringValue("foo")
		v2 := NewStringValue("bar")
		_, err := AddValues(v1, v2)
		assert.Error(t, err)
	})

	t.Run("mismatched types", func(t *testing.T) {
		v1 := NewStringValue("foo")
		v2 := NewIntegerValue(1)
		_, err := AddValues(v1, v2)
		assert.Error(t, err)
	})

	t.Run("unsupported type", func(t *testing.T) {
		v1, _ := NewListValue([]Value{NewIntegerValue(1)})
		v2, _ := NewListValue([]Value{NewIntegerValue(2)})
		_, err := AddValues(v1, v2)
		assert.Error(t, err)
	})
}
