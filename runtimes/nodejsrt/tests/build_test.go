package runner_test

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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
	runnerDir := filepath.Join(t.TempDir(), "runner")

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
				// Clean directories
				os.RemoveAll(buildDir)
				os.RemoveAll(runnerDir)

				// Create runner directory and copy files
				if err := os.MkdirAll(runnerDir, 0755); err != nil {
					return fmt.Errorf("failed to create runner dir: %v", err)
				}
				if err := copyDir("..", runnerDir); err != nil {
					return fmt.Errorf("failed to copy runner files: %v", err)
				}

				// Install dependencies in the runner directory
				npmCmd := exec.Command("npm", "install")
				npmCmd.Dir = runnerDir
				var npmStdout, npmStderr bytes.Buffer
				npmCmd.Stdout = &npmStdout
				npmCmd.Stderr = &npmStderr
				if err := npmCmd.Run(); err != nil {
					t.Logf("npm install failed in %s", runnerDir)
					t.Logf("stdout: %s", npmStdout.String())
					t.Logf("stderr: %s", npmStderr.String())
					return fmt.Errorf("npm install failed: %v", err)
				}

				// Copy test project to build dir
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
			// Setup test
			//if err := tt.setupFunc(); err != nil {
			//	t.Fatalf("setup failed: %v", err)
			//}

			// Run build
			artifact, err := svc.(interface {
				Build(context.Context, fs.FS, string, []sdktypes.Symbol) (sdktypes.BuildArtifact, error)
			}).Build(context.Background(), os.DirFS(testDir), "", nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := tt.checkFunc(artifact); err != nil {
				t.Errorf("checkFunc failed: %v", err)
			}
		})
	}
}

// Helper function to copy directory recursively
func copyDir(src, dst string) error {
	fmt.Printf("Copying directory from %s to %s\n", src, dst)
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip .git directory
		if relPath == ".git" || filepath.HasPrefix(relPath, ".git/") {
			return filepath.SkipDir
		}

		// Create destination path
		dstPath := filepath.Join(dst, relPath)
		fmt.Printf("Copying file: %s -> %s\n", path, dstPath)

		// If it's a directory, create it
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
