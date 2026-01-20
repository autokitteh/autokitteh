package runner_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestInvoicesBuild(t *testing.T) {
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

	// Test setup - start with simple-test project first
	testDir := filepath.Join("..", "examples", "simple-test")
	buildDir := filepath.Join(t.TempDir(), "build")
	runnerDir := filepath.Join(t.TempDir(), "runner")

	// Clean directories
	os.RemoveAll(buildDir)
	os.RemoveAll(runnerDir)

	// Create runner directory and copy files
	if err := os.MkdirAll(runnerDir, 0755); err != nil {
		t.Fatalf("failed to create runner dir: %v", err)
	}
	if err := copyDir(filepath.Join("..", "runner"), runnerDir); err != nil {
		t.Fatalf("failed to copy runner files: %v", err)
	}

	// Install dependencies in the runner directory
	npmCmd := exec.Command("npm", "install")
	npmCmd.Dir = runnerDir
	if err := npmCmd.Run(); err != nil {
		t.Fatalf("npm install failed: %v", err)
	}

	// Copy test project to build dir
	if err := copyDir(testDir, buildDir); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Run build
	artifact, err := svc.(interface {
		Build(context.Context, fs.FS, string, []sdktypes.Symbol) (sdktypes.BuildArtifact, error)
	}).Build(context.Background(), os.DirFS(buildDir), "", nil)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Print build artifacts
	t.Logf("Build Artifact:")
	t.Logf("  Is Valid: %v", artifact.IsValid())

	compiledData := artifact.CompiledData()
	t.Logf("  Compiled Data Keys:")
	for k := range compiledData {
		t.Logf("    - %s", k)
	}

	exports := artifact.Exports()
	t.Logf("  Exports:")
	for _, exp := range exports {
		pb := exp.ToProto()
		t.Logf("    - Symbol: %v", pb.Symbol)
		if pb.Location != nil {
			t.Logf("      Location: %s:%d", pb.Location.Path, pb.Location.Row)
		}
	}
}
