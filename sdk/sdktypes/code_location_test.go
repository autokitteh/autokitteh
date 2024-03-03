package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeLocation(t *testing.T) {
	tests := []struct {
		s          string
		path, name string
		row, col   uint32
		err        bool
	}{
		{
			err: true,
		},
		{
			s:    "meow.kitteh",
			path: "meow.kitteh",
		},
		{
			s:    "meow.kitteh:42",
			path: "meow.kitteh",
			row:  42,
		},
		{
			s:    "meow.kitteh:42.16",
			path: "meow.kitteh",
			row:  42,
			col:  16,
		},
		{
			s:    "meow.kitteh:42.16,Meow",
			path: "meow.kitteh",
			row:  42,
			col:  16,
			name: "Meow",
		},
		{
			s:    "meow.kitteh:42,Meow",
			path: "meow.kitteh",
			row:  42,
			name: "Meow",
		},
		{
			s:    "meow.kitteh:Meow",
			path: "meow.kitteh",
			name: "Meow",
		},
		{
			s:    ":Meow",
			name: "Meow",
		},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			l, err := StrictParseCodeLocation(test.s)

			if test.err {
				assert.Error(t, err)
				assert.Zero(t, l)
				return
			}

			if !assert.NoError(t, err) || !assert.NotNil(t, l) {
				return
			}

			assert.Equal(t, test.path, l.Path())
			assert.Equal(t, test.name, l.Name())

			r, c := l.Row(), l.Col()
			assert.Equal(t, test.row, r)
			assert.Equal(t, test.col, c)

			assert.Equal(t, test.s, l.CanonicalString())
		})
	}
}
