package temporaldevsrv

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.temporal.io/sdk/client"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

// StartDevServer starts a Temporal CLI dev server process. This may download the server if not already downloaded.
// The server binary must exists either in options.ExistingPath or be downloaded to the path specified in options.CachedDownload.DestDir.
func StartDevServer(ctx context.Context, options DevServerOptions) (*DevServer, error) {
	exePath := options.ExistingPath
	if exePath == "" {
		var exists bool
		if exePath, _, exists = GetDownloadInfo(ctx, options.CachedDownload); !exists {
			return nil, sdkerrors.ErrNotFound
		}
	}

	clientOptions := options.clientOptionsOrDefault()

	if clientOptions.HostPort == "" {
		var err error

		// Make sure this is done after downloading to reduce the chance (however slim) that the free port would be used
		// up by the time the download completes.
		clientOptions.HostPort, err = getFreeHostPort()
		if err != nil {
			return nil, err
		}
	}

	host, port, err := net.SplitHostPort(clientOptions.HostPort)
	if err != nil {
		return nil, fmt.Errorf("invalid HostPort: %w", err)
	}

	args := prepareCommand(&options, host, port, clientOptions.Namespace)

	cmd := newCmd(exePath, args...)
	if options.Stdout != nil {
		cmd.Stdout = options.Stdout
	}
	if options.Stderr != nil {
		cmd.Stderr = options.Stderr
	}

	clientOptions.Logger.Info("Starting DevServer", "ExePath", exePath, "Args", args)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed starting: %w", err)
	}

	returnedClient, err := waitServerReady(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	clientOptions.Logger.Info("DevServer ready")
	return &DevServer{
		client:           returnedClient,
		cmd:              cmd,
		frontendHostPort: clientOptions.HostPort,
	}, nil
}

func prepareCommand(options *DevServerOptions, host, port, namespace string) []string {
	args := []string{
		"server",
		"start-dev",
		"--ip", host, "--port", port,
		"--namespace", namespace,
		"--dynamic-config-value", "frontend.enableServerVersionCheck=false",
	}
	if options.LogLevel != "" {
		args = append(args, "--log-level", options.LogLevel)
	}
	if options.LogFormat != "" {
		args = append(args, "--log-format", options.LogFormat)
	}
	if !options.EnableUI {
		args = append(args, "--headless")
	}
	if options.DBFilename != "" {
		args = append(args, "--db-filename", options.DBFilename)
	}
	if options.UIPort != "" {
		args = append(args, "--ui-port", options.UIPort)
	}
	return append(args, options.ExtraArgs...)
}

// waitServerReady repeatedly attempts to dial the server with given options until it is ready or it is time to give up.
// Returns a connected client created using the provided options.
func waitServerReady(ctx context.Context, options client.Options) (client.Client, error) {
	var returnedClient client.Client
	lastErr := retryFor(ctx, 600, 100*time.Millisecond, func() error {
		var err error
		returnedClient, err = client.DialContext(ctx, options)
		return err
	})
	if lastErr != nil {
		return nil, fmt.Errorf("failed connecting after timeout, last error: %w", lastErr)
	}
	return returnedClient, lastErr
}

// retryFor retries some function until it returns nil or runs out of attempts. Wait interval between attempts.
func retryFor(ctx context.Context, maxAttempts int, interval time.Duration, cond func() error) error {
	if maxAttempts < 1 {
		// this is used internally, okay to panic
		panic("maxAttempts should be at least 1")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastErr error
	for range maxAttempts {
		if curE := cond(); curE == nil {
			return nil
		} else {
			lastErr = curE
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Try again after waiting up to interval.
		}
	}
	return lastErr
}
