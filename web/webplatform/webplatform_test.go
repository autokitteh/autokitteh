package webplatform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	fs, v, err := LoadFS()
	require.NoError(t, err)
	require.NotNil(t, fs)

	assert.Regexp(t, `^autokitteh-web-v\d+\.\d+\.\d+$`, v)

	f, err := fs.Open("index.html")
	require.NoError(t, err)
	require.NotNil(t, f)
}
