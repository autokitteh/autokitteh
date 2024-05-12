package temporalclient

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap/zaptest"
)

func TestStartDevServer(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

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
	defer c.Stop(ctx)

	require.NotNil(t, c.client)
	require.NotNil(t, c.srv)
	require.NotNil(t, c.srvLog)

	time.Sleep(100 * time.Millisecond) // Let the server emit some logs

	file, err := os.Open(c.srvLog.Name())
	require.NoError(t, err)
	defer file.Close()
	data, err := io.ReadAll(file)
	require.NoError(t, err)
	require.NotEmpty(t, data)
}
