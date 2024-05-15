package sdkmodule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO: These tests are faaaaaaar from exhaustive.
func TestUnpack(t *testing.T) {
	assert.NoError(t, UnpackArgs(nil, nil))

	var (
		i  int
		st struct {
			X int
		}
	)

	assert.Error(t, UnpackArgs(nil, nil, "i", &i, &st))

	if assert.NoError(t, UnpackArgs([]sdktypes.Value{sdktypes.NewIntegerValue(42)}, nil, "i", &i)) {
		assert.Equal(t, 42, i)
		assert.Zero(t, st)
	}

	var s string

	if assert.NoError(t, UnpackArgs([]sdktypes.Value{sdktypes.NewIntegerValue(64)}, map[string]sdktypes.Value{
		"s": sdktypes.NewStringValue("meow"),
	}, "i", &i, "s=", &s)) {
		assert.Equal(t, 64, i)
		assert.Equal(t, "meow", s)
		assert.Zero(t, st)
	}
}

func TestUnpackFlat(t *testing.T) {
	var (
		i  int
		st struct {
			X int     `json:"x"`
			Y *string `json:"y"`
			S struct {
				Z int `json:"z"`
			} `json:"s"`
			Sptr *struct {
				Z int `json:"z"`
			} `json:"s"`
		}
	)

	assert.NoError(t, UnpackArgs([]sdktypes.Value{
		sdktypes.NewIntegerValue(42),
	}, map[string]sdktypes.Value{
		"x": sdktypes.NewIntegerValue(64),
		"y": sdktypes.NewStringValue("meow"),
		"s": sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
			"z": sdktypes.NewIntegerValue(128),
		}),
	}, "i", &i, &st))

	assert.Equal(t, 42, i)
	assert.Equal(t, 64, st.X)
	assert.Equal(t, "meow", *st.Y)
	assert.Equal(t, 128, st.S.Z)
	assert.Zero(t, st.Sptr)
}
