package temporalclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/xdg"
)

func TestStartDevServer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	xdg.Reload()

	c := &impl{
		cfg: &Config{
			DevServer: testsuite.DevServerOptions{},
		},
		opts: client.Options{},
		done: make(chan struct{}),
		z:    zaptest.NewLogger(t),
	}

	ctx := context.Background()
	err := c.startDevServer(ctx)
	require.NoError(t, err)
	defer func() {
		err := c.Stop(ctx)
		require.NoError(t, err)
	}()

	require.NotNil(t, c.client)
	require.NotNil(t, c.srv)
	require.NotNil(t, c.logFile)
}
