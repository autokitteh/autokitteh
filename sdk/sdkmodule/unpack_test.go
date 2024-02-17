package sdkmodule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO: These tests are faaaaaaar from exhaustive.
func TestUnpack(t *testing.T) {
	assert.NoError(t, UnpackArgs(nil, nil))

	var i int
	assert.Error(t, UnpackArgs(nil, nil, "i", &i))

	if assert.NoError(t, UnpackArgs([]sdktypes.Value{sdktypes.NewIntegerValue(42)}, nil, "i", &i)) {
		assert.Equal(t, 42, i)
	}

	var s string

	if assert.NoError(t, UnpackArgs([]sdktypes.Value{sdktypes.NewIntegerValue(64)}, map[string]sdktypes.Value{
		"s": sdktypes.NewStringValue("meow"),
	}, "i", &i, "s=", &s)) {
		assert.Equal(t, 64, i)
		assert.Equal(t, "meow", s)
	}
}
