package webplatform

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestLoad(t *testing.T) {
	fs, v, err := LoadFS(zaptest.NewLogger(t))
	if errors.Is(err, sdkerrors.ErrNotFound) {
		t.Skip("no web platform distribution found")
	}

	require.NoError(t, err)
	require.NotNil(t, fs)

	t.Logf("version: %s", v)

	assert.Regexp(t, `^\d+\.\d+\.\d+$`, v)

	f, err := fs.Open("index.html")
	require.NoError(t, err)
	require.NotNil(t, f)
}
