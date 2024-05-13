package temporalclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap/zaptest"
)

// FIXME: Disabled due to ENG-836, until there's an official release of the Temporal SDK.
func DisabledTestStartDevServer(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)
	xdg.Reload()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	cfg := &Config{
		DevServer: testsuite.DevServerOptions{},
	}
	opts := client.Options{}

	c := &impl{
		done: make(chan struct{}),
		z:    logger,
	}
	err := c.startDevServer(ctx, cfg, opts)
	require.NoError(t, err)
	defer func() {
		err := c.Stop(ctx)
		require.NoError(t, err)
	}()

	require.NotNil(t, c.client)
	require.NotNil(t, c.srv)
	require.NotNil(t, c.srvLog)
}
