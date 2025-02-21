package sessions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{
			"python_stdlib",
			`File "/opt/hostedtoolcache/Python/3.12.8/x64/lib/python3.12/concurrent/futures/_base.py", line 401, in __get_result`,
			`py-lib/concurrent/futures/_base.py, line XXX, in __get_result`,
		},
		{
			"user_code",
			"/tmp/ak-user-2767870919/main.py:6.1,main",
			"main.py:6.1,main",
		},
		{
			"ak_runner_main",
			"/tmp/ak-runner-2767870918/main.py:6.1,main",
			"   ak-runner",
		},
		{
			"ak_runner_call",
			"runner/main.py:6.1,main, in _call",
			"   ak-runner",
		},
		{
			"error",
			"ERROR: bad token",
			"ERROR: bad token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePath(tt.path)
			require.Equal(t, tt.want, got)
		})
	}
}
