package sdktypes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestEventFilter(t *testing.T) {
	e := kittehs.Must1(sdktypes.EventFromProto(
		&sdktypes.EventPB{
			Data: map[string]*valuev1.Value{
				"foo": sdktypes.NewStringValue("meow").ToProto(),
			},
		},
	))

	matches, err := e.Matches("")
	if assert.NoError(t, err) {
		assert.True(t, matches)
	}

	matches, err = e.Matches("data.foo == 'meow'")
	if assert.NoError(t, err) {
		assert.True(t, matches)
	}

	matches, err = e.Matches("has(data.foo) && data.foo == 'meow'")
	if assert.NoError(t, err) {
		assert.True(t, matches)
	}

	matches, err = e.Matches("has(data.bar) && data.bar == 'hiss'")
	if assert.NoError(t, err) {
		assert.False(t, matches)
	}
}
