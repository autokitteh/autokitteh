package webplatform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	fs, err := LoadFS()
	require.NoError(t, err)
	require.NotNil(t, fs)

	f, err := fs.Open("index.html")
	require.NoError(t, err)
	require.NotNil(t, f)
}
