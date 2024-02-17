package sdktypes

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestParseSymbol(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{
			in: "meow",
		},
		{
			in: "_meow",
		},
		{
			in: "x",
		},
		{
			in: "_",
		},
		{
			in: "plan9",
		},
		{
			in:      "9plan",
			wantErr: true,
		},
		{
			in:      "x ",
			wantErr: true,
		},
		{
			in:      "_ ",
			wantErr: true,
		},
		{
			in:      "@meow",
			wantErr: true,
		},
		{
			in:      " meow",
			wantErr: true,
		},
		{
			in:      "me@ow",
			wantErr: true,
		},
		{
			in:      "1meow",
			wantErr: true,
		},
		{
			in:      "@meow",
			wantErr: true,
		},
		{
			in:      "woof ",
			wantErr: true,
		},
		{
			in:      " ",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			check := func(f func(string) (Symbol, error)) {
				s, err := f(test.in)

				if test.wantErr {
					assert.Nil(t, s)
					if assert.Error(t, err) {
						assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
					}
					return
				}

				assert.NoError(t, err)
				if !assert.NotNil(t, s) {
					return
				}

				assert.Equal(t, test.in, s.String())
			}

			check(ParseSymbol)
			check(StrictParseSymbol)
		})
	}

	t.Run("parse empty symbol allowed", func(t *testing.T) {
		s, err := ParseSymbol("")
		assert.NoError(t, err)
		assert.Nil(t, s)
	})

	t.Run("strict parse empty symbol not allowed", func(t *testing.T) {
		s, err := StrictParseSymbol("")
		if assert.Error(t, err) {
			assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
		}
		assert.Nil(t, s)
	})
}
