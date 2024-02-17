package sdktypes

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type comparableID interface {
	comparable
	ID
}
type idFuncs[SpecificID comparableID, Traits idTraits] struct {
	traits             Traits
	New                func() SpecificID
	Parse, StrictParse func(string) (SpecificID, error)
	ParseIDOrName      func(string) (Name, SpecificID, error)
}

func testID[SpecificID comparableID, Traits idTraits](
	t *testing.T, fns idFuncs[SpecificID, Traits],
) {
	var zero SpecificID

	t.Run("unique", func(t *testing.T) {
		const n = 100
		ids := make(map[string]bool, n)
		for i := 0; i < n; i++ {
			id := fns.New()
			if assert.NotNil(t, id) {
				assert.False(t, ids[id.String()])
			}
			ids[id.String()] = true
		}
	})

	t.Run("valid", func(t *testing.T) {
		id := fns.New()
		assert.Equal(t, fns.traits.Kind(), id.Kind())
		assert.True(t, IsValidID(id.String()))
	})

	t.Run("parse", func(t *testing.T) {
		tests := []struct {
			in  string // %s will be replaced by kind.
			err bool
		}{
			{
				in: "%s:00000000000000000000000000000000",
			},
			{
				in:  ":00000000000000000000000000000000",
				err: true,
			},
			{
				in:  "%s:",
				err: true,
			},
			{
				in:  "%s:0000000000000000000000000000meow",
				err: true,
			},
			{
				in:  "%s:0",
				err: true,
			},
			{
				in:  "%s00000000000000000000000000000000",
				err: true,
			},
		}

		for _, test := range tests {
			kind := fns.traits.Kind()
			in := fmt.Sprintf(test.in, kind)

			t.Run(in, func(t *testing.T) {
				check := func(p func(string) (SpecificID, error)) {
					id, err := fns.Parse(in)

					if test.err {
						assert.Error(t, err)
						assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
						assert.Nil(t, id)
						return
					}

					if !assert.NoError(t, err) || !assert.NotNil(t, id) {
						return
					}

					assert.Equal(t, in, id.String())
					assert.Equal(t, kind, id.Kind())
				}

				check(fns.Parse)
				check(fns.StrictParse)

				id, err := fns.Parse("")
				assert.NoError(t, err)
				assert.Nil(t, id)

				id, err = fns.StrictParse("")
				if assert.Error(t, err) {
					assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument))
				}
				assert.Nil(t, id)
			})
		}
	})

	t.Run("parse_id_or_handle", func(t *testing.T) {
		tests := []struct {
			in   string // %s will be replaced by kind.
			isID bool   // else, handle.
			err  bool
		}{
			{
				err: true,
			},
			{
				in: "meow",
			},
			{
				in:   "%s:00000000000000000000000000000000",
				isID: true,
			},
			{
				in:   "x:00000000000000000000000000000000",
				isID: true,
				err:  true,
			},
			{
				in:  "meow woof",
				err: true,
			},
		}

		for _, test := range tests {
			in, kind := test.in, fns.traits.Kind()

			if strings.Contains(in, "%s") {
				in = fmt.Sprintf(in, kind)
			}

			t.Run(in, func(t *testing.T) {
				h, id, err := fns.ParseIDOrName(in)

				if test.err {
					assert.Error(t, err)
					assert.True(t, errors.Is(err, sdkerrors.ErrInvalidArgument), err)
					assert.False(t, IsValidName(in))
					// NOPE: assert.False(t, IsValidID(in)) (might be a valid id, but with the wrong kind)
					assert.Nil(t, id)
					assert.Nil(t, h)
					return
				}

				if !assert.NoError(t, err) ||
					!assert.False(t, id == zero && h == nil) ||
					!assert.False(t, id != zero && h != nil) {
					return
				}

				if test.isID {
					assert.True(t, IsID(in))
					assert.False(t, IsName(in))

					if assert.NotNil(t, id) {
						assert.Equal(t, in, id.String())
						assert.Equal(t, kind, id.Kind())
						assert.Nil(t, h)
					}

					return
				}

				assert.Nil(t, id)

				assert.False(t, IsID(in))
				assert.True(t, IsName(in))

				if !assert.NotNil(t, h) {
					return
				}

				assert.Equal(t, in, h.String())
			})
		}
	})
}
