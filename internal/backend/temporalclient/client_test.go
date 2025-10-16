package temporalclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/temporaldevsrv"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

func TestStartDevServer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	xdg.Reload()

	c := &impl{
		cfg: &Config{
			DevServer:                   temporaldevsrv.DevServerOptions{},
			DevServerStartMaxAttempts:   3,
			DevServerStartRetryInterval: time.Second,
			DevServerStartTimeout:       time.Second * 5,
		},
		opts: client.Options{},
		done: make(chan struct{}),
		l:    zaptest.NewLogger(t),
	}

	ctx := t.Context()
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
