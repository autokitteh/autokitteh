package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestLocalRunner(t *testing.T) {
	t.Run("Basic Project Setup and Start", func(t *testing.T) {
		// Setup logger
		log := zap.NewExample()
		defer log.Sync()

		// Setup server and configure local runner manager
		mux := &http.ServeMux{}
		nodejsrt.ConfigureWorkerGRPCHandler(log, mux)
		listener, err := setupServer(t, log)
		require.NoError(t, err)
		defer listener.Close()

		// Create config
		cfg := &nodejsrt.Config{
			RunnerType:    "local",
			LogBuildCode:  true,
			LogRunnerCode: true,
		}

		// Create NodeJS runtime
		rt, err := nodejsrt.New(cfg, log, func() string {
			port := listener.Addr().(*net.TCPAddr).Port
			return fmt.Sprintf("localhost:%d", port)
		})
		require.NoError(t, err)

		// Create runtime service
		svc, err := rt.New()
		require.NoError(t, err)

		// Create test project files
		files := map[string][]byte{
			"package.json": []byte(`{
				"name": "test-project",
				"version": "1.0.0",
				"dependencies": {
					"typescript": "^5.0.0"
				}
			}`),
			"tsconfig.json": []byte(`{
				"compilerOptions": {
					"target": "ES2020",
					"module": "commonjs",
					"strict": true,
					"esModuleInterop": true,
					"skipLibCheck": true,
					"forceConsistentCasingInFileNames": true
				}
			}`),
			"index.ts": []byte(`
				export function hello() {
					return "Hello, World!";
				}
			`),
		}

		// Create tar data
		tarData := createTestTar(t, files)

		// Create run ID and session ID
		runID := sdktypes.NewRunID()
		sessionID := sdktypes.NewSessionID()

		// Create callbacks
		cbs := &sdkservices.RunCallbacks{
			Call: func(ctx context.Context, rid sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
				return sdktypes.Nothing, nil
			},
			Load: func(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
				return map[string]sdktypes.Value{}, nil
			},
			Print: func(ctx context.Context, rid sdktypes.RunID, text string) error {
				return nil
			},
		}

		// Run the service
		run, err := svc.Run(
			context.Background(),
			runID,
			sessionID,
			"index.ts",
			map[string][]byte{"index.ts": tarData},
			map[string]sdktypes.Value{},
			cbs,
		)
		require.NoError(t, err)

		// Verify exports
		exports := run.Values()
		require.NotEmpty(t, exports)
		require.Greater(t, len(exports), 0)
	})
}

// Helper function to create a test tar file
func createTestTar(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "test-tar-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	for name, content := range files {
		fullPath := filepath.Join(tmpDir, name)
		dir := filepath.Dir(fullPath)

		// Ensure directory exists
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		// Write file
		err = os.WriteFile(fullPath, content, 0644)
		require.NoError(t, err)
	}

	// Create tar from directory
	fsys := os.DirFS(tmpDir)
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	defer w.Close()

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return err
		}

		// Skip symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = path

		// Write header
		if err := w.WriteHeader(header); err != nil {
			return err
		}

		// If it's a directory, we're done
		if info.IsDir() {
			return nil
		}

		// Copy file content
		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(w, file)
		return err
	})
	require.NoError(t, err)

	return buf.Bytes()
}

// Helper function to setup the server
func setupServer(t *testing.T, log *zap.Logger) (net.Listener, error) {
	t.Helper()

	mux := &http.ServeMux{}
	nodejsrt.ConfigureWorkerGRPCHandler(log, mux)
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	// Create config for local runner
	cfg := &nodejsrt.Config{
		RunnerType:    "local",
		LogRunnerCode: true,
	}

	// Create NodeJS runtime which will configure the local runner manager
	_, err = nodejsrt.New(cfg, log, func() string {
		port := listener.Addr().(*net.TCPAddr).Port
		return fmt.Sprintf("localhost:%d", port)
	})
	if err != nil {
		return nil, fmt.Errorf("configure local runner manager: %w", err)
	}

	errChan := make(chan error)
	go func() {
		if err := http.Serve(listener, h2c.NewHandler(mux, &http2.Server{})); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			return
		}
	}()

	select {
	case err := <-errChan:
		return nil, err
	case <-time.After(2 * time.Second): // give server time to start
		return listener, nil
	}
}
