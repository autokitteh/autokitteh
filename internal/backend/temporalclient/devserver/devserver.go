// Forked from https://pkg.go.dev/go.temporal.io/sdk@v1.26.1/testsuite
// do disable logging to stdout/stderr

// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package devserver

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
)

var SDKVersion = "1.26.0"

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
	// Log format - defaults to "pretty".
	LogFormat string
	// Log level - defaults to "warn".
	LogLevel string
	// Additional arguments to the dev server.
	ExtraArgs []string
}

// Temporal CLI based DevServer
type DevServer struct {
	cmd              *exec.Cmd
	client           client.Client
	frontendHostPort string
}

// StartDevServer starts a Temporal CLI dev server process. This may download the server if not already downloaded.
func StartDevServer(ctx context.Context, options DevServerOptions) (*DevServer, error) {
	clientOptions := options.clientOptionsOrDefault()

	exePath, err := downloadIfNeeded(ctx, &options, clientOptions.Logger)
	if err != nil {
		return nil, err
	}

	if clientOptions.HostPort == "" {
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
	return append(args, options.ExtraArgs...)
}

func downloadIfNeeded(ctx context.Context, options *DevServerOptions, logger log.Logger) (string, error) {
	if options.ExistingPath != "" {
		return options.ExistingPath, nil
	}
	version := options.CachedDownload.Version
	if version == "" {
		version = "default"
	}
	destDir := options.CachedDownload.DestDir
	if destDir == "" {
		destDir = os.TempDir()
	}
	var exePath string
	// Build path based on version and check if already present
	if version == "default" {
		exePath = filepath.Join(destDir, "temporal-cli-go-sdk-"+SDKVersion)
	} else {
		exePath = filepath.Join(destDir, "temporal-cli-"+version)
	}
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	if _, err := os.Stat(exePath); err == nil {
		return exePath, nil
	}

	client := &http.Client{}

	// Build info URL
	platform := runtime.GOOS
	if platform != "windows" && platform != "darwin" && platform != "linux" {
		return "", fmt.Errorf("unsupported platform %v", platform)
	}
	arch := runtime.GOARCH
	if arch != "amd64" && arch != "arm64" {
		return "", fmt.Errorf("unsupported architecture %v", arch)
	}
	infoURL := fmt.Sprintf("https://temporal.download/cli/%v?platform=%v&arch=%v&sdk-name=sdk-go&sdk-version=%v", url.QueryEscape(version), platform, arch, SDKVersion)

	// Get info
	info := struct {
		ArchiveURL    string `json:"archiveUrl"`
		FileToExtract string `json:"fileToExtract"`
	}{}
	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed preparing request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed fetching info: %w", err)
	}
	b, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		logger.Warn("Failed to close response body: %v", closeErr)
	}
	if err != nil {
		return "", fmt.Errorf("failed fetching info body: %w", err)
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed fetching info, status: %v, body: %s", resp.Status, b)
	} else if err = json.Unmarshal(b, &info); err != nil {
		return "", fmt.Errorf("failed unmarshaling info: %w", err)
	}

	// Download and extract
	logger.Info("Downloading temporal CLI", "Url", info.ArchiveURL, "ExePath", exePath)
	req, err = http.NewRequestWithContext(ctx, "GET", info.ArchiveURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed preparing request: %w", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed downloading: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body: %v", closeErr)
		}
	}()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed downloading, status: %v", resp.Status)
	}
	// We want to download to a temporary file then rename. A better system-wide
	// atomic downloader would use a common temp file and check whether it exists
	// and wait on it, but doing multiple downloads in racy situations is
	// good/simple enough for now.
	// Note that we don't use os.TempDir here, instead we use the user provided destination directory which is
	// guaranteed to make the rename atomic.
	f, err := os.CreateTemp(destDir, "temporal-cli-downloading-")
	if err != nil {
		return "", fmt.Errorf("failed creating temp file: %w", err)
	}
	if strings.HasSuffix(info.ArchiveURL, ".tar.gz") {
		err = extractTarball(resp.Body, info.FileToExtract, f)
	} else if strings.HasSuffix(info.ArchiveURL, ".zip") {
		err = extractZip(resp.Body, info.FileToExtract, f)
	} else {
		err = fmt.Errorf("unrecognized file extension on %v", info.ArchiveURL)
	}
	closeErr := f.Close()
	if err != nil {
		return "", err
	} else if closeErr != nil {
		return "", fmt.Errorf("failed to close temp file: %w", closeErr)
	}
	// Chmod it if not Windows
	if runtime.GOOS != "windows" {
		if err := os.Chmod(f.Name(), 0755); err != nil {
			return "", fmt.Errorf("failed chmod'ing file: %w", err)
		}
	}
	if err = os.Rename(f.Name(), exePath); err != nil {
		return "", fmt.Errorf("failed moving file: %w", err)
	}
	return exePath, nil
}

func (opts *DevServerOptions) clientOptionsOrDefault() client.Options {
	var out client.Options
	if opts.ClientOptions != nil {
		// Shallow copy the client options since we intend to overwrite some fields.
		out = *opts.ClientOptions
	}
	if out.Logger == nil {
		// FIXME: Get logger from autokitteh
		out.Logger = log.NewStructuredLogger(slog.Default())
	}
	if out.Namespace == "" {
		out.Namespace = "default"
	}
	return out
}

func extractTarball(r io.Reader, toExtract string, w io.Writer) error {
	r, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tarRead := tar.NewReader(r)
	for {
		h, err := tarRead.Next()
		if err != nil {
			// This can be EOF which means we never found our file
			return err
		} else if h.Name == toExtract {
			_, err = io.Copy(w, tarRead)
			return err
		}
	}
}

func extractZip(r io.Reader, toExtract string, w io.Writer) error {
	// Instead of using a third party zip streamer, and since Go stdlib doesn't
	// support streaming read, we'll just put the entire archive in memory for now
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	zipRead, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}
	for _, file := range zipRead.File {
		if file.Name == toExtract {
			r, err := file.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(w, r)
			return err
		}
	}
	return fmt.Errorf("could not find file in zip archive")
}

// waitServerReady repeatedly attempts to dial the server with given options until it is ready or it is time to give up.
// Returns a connected client created using the provided options.
func waitServerReady(ctx context.Context, options client.Options) (client.Client, error) {
	var returnedClient client.Client
	lastErr := retryFor(600, 100*time.Millisecond, func() error {
		var err error
		returnedClient, err = client.Dial(options)
		return err
	})
	if lastErr != nil {
		return nil, fmt.Errorf("failed connecting after timeout, last error: %w", lastErr)
	}
	return returnedClient, lastErr
}

// retryFor retries some function until it returns nil or runs out of attempts. Wait interval between attempts.
func retryFor(maxAttempts int, interval time.Duration, cond func() error) error {
	if maxAttempts < 1 {
		// this is used internally, okay to panic
		panic("maxAttempts should be at least 1")
	}
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if curE := cond(); curE == nil {
			return nil
		} else {
			lastErr = curE
		}
		time.Sleep(interval)
	}
	return lastErr
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
