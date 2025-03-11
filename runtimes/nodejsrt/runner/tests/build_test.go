package runner_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestBuild(t *testing.T) {
	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Create config
	cfg := &nodejsrt.Config{
		RunnerType:   "local",
		LogBuildCode: true,
	}

	// Create NodeJS runtime
	rt, err := nodejsrt.New(cfg, logger, func() string { return "localhost:0" })
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	// Create runtime service
	svc, err := rt.New()
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Test setup
	testDir := "fixtures/test_project"
	buildDir := filepath.Join(t.TempDir(), "build")

	// Test cases
	tests := []struct {
		name      string
		setupFunc func() error
		wantErr   bool
		checkFunc func(sdktypes.BuildArtifact) error
	}{
		{
			name: "successful build",
			setupFunc: func() error {
				// Copy test project to temp dir
				return copyDir(testDir, buildDir)
			},
			wantErr: false,
			checkFunc: func(artifact sdktypes.BuildArtifact) error {
				// Check if we got valid exports
				exports := artifact.Exports()
				if len(exports) == 0 {
					return nil
				}

				// Verify expected exports exist
				expectedExports := []string{"helloWorld", "add", "fetchData", "remoteFunction"}
				for _, exp := range expectedExports {
					found := false
					for _, export := range exports {
						pb := export.ToProto()
						if pb.Symbol == exp {
							found = true
							break
						}
					}
					if !found {
						return nil
					}
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean build directory
			os.RemoveAll(buildDir)

			// Setup test
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Run build
			artifact, err := svc.(interface {
				Build(context.Context, fs.FS, string, []sdktypes.Symbol) (sdktypes.BuildArtifact, error)
			}).Build(context.Background(), os.DirFS(buildDir), "", nil)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Run additional checks
			if tt.checkFunc != nil {
				if err := tt.checkFunc(artifact); err != nil {
					t.Errorf("check failed: %v", err)
				}
			}
		})
	}
}

// Helper function to copy directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Get destination path
		dstPath := filepath.Join(dst, relPath)

		// Create directory if needed
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
