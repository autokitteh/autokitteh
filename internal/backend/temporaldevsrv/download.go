package temporaldevsrv

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/xdg"
)

func GetDownloadInfo(ctx context.Context, desired CachedDownload) (exePath string, cached CachedDownload, exists bool) {
	cached = desired

	if cached.Version == "" {
		cached.Version = "default"
	}

	if cached.DestDir == "" {
		cached.DestDir = xdg.CacheHomeDir()
	}

	// Build path based on version and check if already present
	if cached.Version == "default" {
		exePath = filepath.Join(cached.DestDir, "temporal-cli-go-sdk-"+temporal.SDKVersion)
	} else {
		exePath = filepath.Join(cached.DestDir, "temporal-cli-"+cached.Version)
	}

	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}

	_, err := os.Stat(exePath)
	exists = err == nil

	return
}

func Download(ctx context.Context, desired CachedDownload, l *zap.Logger) (string, error) {
	exePath, cached, exists := GetDownloadInfo(ctx, desired)

	l = l.With(
		zap.String("dest_dir", cached.DestDir),
		zap.String("version", cached.Version),
		zap.String("exe_path", exePath),
	)

	if exists {
		l.Info("server already downloaded")
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
	infoURL := fmt.Sprintf("https://temporal.download/cli/%v?platform=%v&arch=%v&sdk-name=sdk-go&sdk-version=%v", url.QueryEscape(cached.Version), platform, arch, temporal.SDKVersion)

	// Get info
	info := struct {
		ArchiveURL    string `json:"archiveUrl"`
		FileToExtract string `json:"fileToExtract"`
	}{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed preparing request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed fetching info: %w", err)
	}
	b, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		l.Warn("Failed to close response body", zap.Error(closeErr))
	}
	if err != nil {
		return "", fmt.Errorf("failed fetching info body: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed fetching info, status: %v, body: %s", resp.Status, b)
	} else if err = json.Unmarshal(b, &info); err != nil {
		return "", fmt.Errorf("failed unmarshalling info: %w", err)
	}

	// Download and extract
	l.Info("downloading", zap.String("url", info.ArchiveURL))

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, info.ArchiveURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed preparing request: %w", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed downloading: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			l.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed downloading, status: %v", resp.Status)
	}
	// We want to download to a temporary file then rename. A better system-wide
	// atomic downloader would use a common temp file and check whether it exists
	// and wait on it, but doing multiple downloads in racy situations is
	// good/simple enough for now.
	// Note that we don't use os.TempDir here, instead we use the user provided destination directory which is
	// guaranteed to make the rename atomic.
	f, err := os.CreateTemp(cached.DestDir, "temporal-cli-downloading-")
	if err != nil {
		return "", fmt.Errorf("failed creating temp file: %w", err)
	}

	switch {
	case strings.HasSuffix(info.ArchiveURL, ".tar.gz"):
		err = extractTarball(resp.Body, info.FileToExtract, f)
	case strings.HasSuffix(info.ArchiveURL, ".zip"):
		err = extractZip(resp.Body, info.FileToExtract, f)
	default:
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
		if err := os.Chmod(f.Name(), 0o755); err != nil {
			return "", fmt.Errorf("failed chmod'ing file: %w", err)
		}
	}
	if err = os.Rename(f.Name(), exePath); err != nil {
		return "", fmt.Errorf("failed moving file: %w", err)
	}

	l.Info("downloaded", zap.String("path", exePath))

	return exePath, nil
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
			return fmt.Errorf("read %q: %w", toExtract, err)
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
	return errors.New("could not find file in zip archive")
}
