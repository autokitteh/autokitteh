package temporalclient

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporaldevsrv"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

func (c *impl) startDevServer(ctx context.Context) error {
	exePath, err := temporaldevsrv.Download(ctx, c.cfg.DevServer.CachedDownload, c.l)
	if err != nil {
		return fmt.Errorf("download temporal dev server: %w", err)
	}

	logPath := path.Join(xdg.DataHomeDir(), "temporal_dev.log")
	c.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open Temporal dev server log file: %w", err)
	}

	devSrvCfg := c.cfg.DevServer
	devSrvCfg.ExistingPath = exePath
	devSrvCfg.ClientOptions = &c.opts
	devSrvCfg.Stderr = c.logFile
	devSrvCfg.Stdout = c.logFile

	for i := range c.cfg.DevServerStartMaxAttempts {
		if c.srv != nil {
			break
		}

		l := c.l.With(zap.Int("attempt", i))

		l.Info("starting temporal dev server")

		if i > 0 {
			select {
			case <-time.After(c.cfg.DevServerStartRetryInterval):
				// nop
			case <-ctx.Done():
				return fmt.Errorf("context done: %w", ctx.Err())
			}
		}

		startCtx := ctx
		done := func() {}

		if c.cfg.DevServerStartTimeout != 0 {
			startCtx, done = context.WithTimeout(ctx, c.cfg.DevServerStartTimeout)
		}

		c.srv, err = temporaldevsrv.StartDevServer(startCtx, devSrvCfg)
		done()

		if err != nil {
			l.Error("Failed to starting temporal dev server. Check temporal log for further info.", zap.Error(err), zap.String("log_path", logPath))
			continue
		}

		// Give additional time to the server to start up, create namespace, etc.
		time.Sleep(c.cfg.DevServerStartWaitTime)
	}

	if err != nil {
		return fmt.Errorf("start Temporal dev server: %w", err)
	}

	c.l.Info("started temporal dev server", zap.String("address", c.srv.FrontendHostPort()))

	if c.client != nil {
		c.client.Close()
	}

	c.client = c.srv.Client()

	return nil
}
