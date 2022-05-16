//go:build unit

package idgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSequentialPerPrefix(t *testing.T) {
	g := NewSequentialPerPrefix(0)

	require.Equal(t, "X1", g("X"))
	require.Equal(t, "Y1", g("Y"))
	require.Equal(t, "X2", g("X"))
	require.Equal(t, "X3", g("X"))
	require.Equal(t, "Y2", g("Y"))
}
