package sessions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var normCases = []struct {
	path   string
	normed string
}{
	{"", ""},
	{"/usr/lib/python3.12/concurrent/futures/_base.py", "py/concurrent/futures/_base.py"},
	{"/tmp/ak-user-2767870919/main.py:6.1,main", "main.py:6.1,main"},
	{"/tmp/ak-runner-2767870918/main.py:6.1,main", ""},
	{"runner/main.py:6.1,main, in _call", ""},
	{"ERROR: bad token", "ERROR: bad token"},
}

func Test_normalizePath(t *testing.T) {
	for _, c := range normCases {
		t.Run(c.path, func(t *testing.T) {
			normed := normalizePath(c.path)
			require.Equal(t, c.normed, normed)
		})
	}
}
