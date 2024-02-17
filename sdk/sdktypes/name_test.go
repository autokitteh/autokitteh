package sdktypes

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestName(t *testing.T) {
	tests := []struct {
		title  string
		in     string
		err    bool
		isName bool // assumed true for non-errors. checked only on errors.
	}{
		{
			title: "valid",
			in:    "meow",
		},
		{
			title:  "invalid",
			in:     "me ow",
			err:    true,
			isName: true,
		},
		{
			title:  "all spaces",
			in:     " ",
			err:    true,
			isName: true,
		},
		{
			title: "numbers",
			in:    "m123",
		},
		{
			title: "just numbers",
			in:    "123",
		},
		{
			title:  "nonalphanum",
			in:     "meow.hiss",
			err:    true,
			isName: true,
		},
	}

	check := func(t *testing.T, f func(string) (Name, error), in string, isErr bool) {
		h, err := f(in)
		if isErr {
			assert.Nil(t, h)

			if assert.NotNil(t, err) {
				assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
			}

			assert.False(t, IsValidName(in))

			return
		}

		if !assert.Nil(t, err) {
			return
		}

		assert.Equal(t, in, h.String())
		assert.True(t, IsValidName(in))
		assert.True(t, IsName(in))
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			check(t, ParseName, test.in, test.err)
			check(t, StrictParseName, test.in, test.err)

			if test.err {
				assert.Equal(t, test.isName, IsName(test.in))
			}
		})
	}

	assert.False(t, IsName(""))
	assert.False(t, IsValidName(""))

	h, err := ParseName("")
	assert.NoError(t, err)
	assert.Nil(t, h)

	h, err = StrictParseName("")
	if assert.Error(t, err) {
		assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
	}
	assert.Nil(t, h)
}
