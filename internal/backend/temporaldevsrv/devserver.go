package temporaldevsrv

import (
	"io"
	"os/exec"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"
)

// Cached download of the dev server.
type CachedDownload struct {
	// Which version to download, by default the latest version compatible with the SDK will be downloaded.
	// Acceptable values are specific release versions (e.g v0.3.0), "default", and "latest".
	Version string
	// Destination directory or the user temp directory if unset.
	DestDir string
}

// Configuration for the dev server.
type DevServerOptions struct {
	// Existing path on the filesystem for the executable.
	ExistingPath string
	// Download the executable if not already there.
	CachedDownload CachedDownload
	// Client options used to create a client for the dev server.
	// The provided Namespace or the "default" namespace is automatically registered on startup.
	// If HostPort is provided, the host and port will be used to bind the server, otherwise the server will bind to
	// localhost and obtain a free port.
	ClientOptions *client.Options
	// SQLite DB filename if persisting or non-persistent if none.
	DBFilename string
	// Whether to enable the UI.
	EnableUI bool
	// Override UI port if EnableUI is true.
	// If not provided, a free port will be used.
	UIPort string
	// Log format - defaults to "pretty".
	LogFormat string
	// Log level - defaults to "warn".
	LogLevel string
	// Additional arguments to the dev server.
	ExtraArgs []string
	// Where to redirect stdout and stderr, if nil they will be redirected to the current process.
	Stdout io.Writer
	Stderr io.Writer
}

// Temporal CLI based DevServer
type DevServer struct {
	cmd              *exec.Cmd
	client           client.Client
	frontendHostPort string
}

func (opts *DevServerOptions) clientOptionsOrDefault() (out client.Options) {
	if opts.ClientOptions != nil {
		// Shallow copy the client options since we intend to overwrite some fields.
		out = *opts.ClientOptions
	}

	if out.Logger == nil {
		out.Logger = logur.LoggerToKV(zapadapter.New(zap.NewNop()))
	}

	if out.Namespace == "" {
		out.Namespace = "default"
	}

	return out
}

// Stop the running server and wait for shutdown to complete. Error is propagated from server shutdown.
func (s *DevServer) Stop() error {
	if err := sendInterrupt(s.cmd.Process); err != nil {
		return err
	}
	return s.cmd.Wait()
}

// Get a connected client, configured to work with the dev server.
func (s *DevServer) Client() client.Client {
	return s.client
}

// FrontendHostPort returns the host:port for this server.
func (s *DevServer) FrontendHostPort() string {
	return s.frontendHostPort
}
