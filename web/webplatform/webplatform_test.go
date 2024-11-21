package webplatform

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestLoad(t *testing.T) {
	fs, v, err := LoadFS(zap.NewNop())
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

func TestVersion(t *testing.T) {
	l, _ := zap.NewDevelopment() // if test fails, will show detailed version error.

	fs, v, err := LoadFS(l)
	if errors.Is(err, sdkerrors.ErrNotFound) {
		t.Skip("no web platform distribution found")
	}

	require.NoError(t, err)
	require.NotNil(t, fs)

	ok, err := ensureVersion(l, v)
	if assert.NoError(t, err) {
		assert.True(t, ok, "version mismatch")
	}
}
