//go:build unit

package apiprogram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	tests := []struct {
		n                 string
		p                 string
		scheme, path, ver string
		err               bool
	}{
		{
			n:   "empty",
			err: true,
		},
		{
			n:    "dot",
			p:    ".",
			path: ".",
		},
		{
			n:    "p",
			p:    "meow",
			path: "meow",
		},
		{
			n:      "sp",
			p:      "github:hello/meow",
			scheme: "github",
			path:   "hello/meow",
		},
		{
			n:      "spv",
			p:      "github:hello/meow#main",
			scheme: "github",
			path:   "hello/meow",
			ver:    "main",
		},
		{
			n:    "pv",
			p:    "hello/meow#main",
			path: "hello/meow",
			ver:  "main",
		},
		{
			n:   "s",
			p:   "meow:",
			err: true,
		},
		{
			n:   "sv",
			p:   "meow:#master",
			err: true,
		},
		{
			n:   "v",
			p:   "#master",
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			p, err := ParsePathString(test.p)

			if test.err {
				assert.Error(t, err)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.EqualValues(t, test.scheme, p.Scheme())
			assert.EqualValues(t, test.path, p.Path())
			assert.EqualValues(t, test.ver, p.Version())
		})
	}
}
