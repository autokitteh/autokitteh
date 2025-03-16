package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var runnerPath = filepath.Join("..", "..", "runtimes", "pythonrt", "runner")

// AKPath returns the absolute path to the AK binary file. It also checks
// that the current working directory is the calling test's directory, and
// the AK binary file exists in the "bin" directory in the repository's root.
func AKPath(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get current working directory:", err)
	}

	path, err := filepath.Abs(filepath.Join(wd, "..", "..", "bin", "ak"))
	if err != nil {
		t.Fatal("failed to construct AK path:", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatal("failed to get AK file info:", err)
	}

	return path
}

// SwitchToTempDir creates a temporary directory, for test isolation.
func SwitchToTempDir(t *testing.T, venvPath string) string {
	tmpPath := t.TempDir()

	t.Chdir(tmpPath)

	// Don't use the user's "config.yaml" file, it may violate isolation
	// by forcing tests to use shared and/or persistent resources.
	t.Setenv("XDG_CONFIG_HOME", tmpPath)
	t.Setenv("XDG_DATA_HOME", tmpPath)

	// Warm up a Python virtual environment as the AK server's runner (i.e. create
	// a symbolic link from the test suite's reusable Python virtual environment,
	// created by [CreatePythonVenv], to this new temporary directory).
	if err := os.Mkdir("autokitteh", 0o755); err != nil {
		t.Fatal("failed to create AK config & data directory:", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("autokitteh"); err != nil {
			t.Error("failed to remove temporary AK config & data directory:", err)
		}
	})
	if err := os.Symlink(venvPath, filepath.Join("autokitteh", "venv")); err != nil {
		t.Fatal("failed to link Python venv in temporary directory:", err)
	}

	return tmpPath
}

// createPythonVenv creates a reusable Python virtual environment (assumed to
// be in the AK Python runner directory) for all the test cases, to avoid
// AK server startup delays. This is removed during test-suite cleanup.
func CreatePythonVenv(t *testing.T) string {
	// https://docs.astral.sh/uv/reference/cli/#uv-sync
	cmd := exec.Command("uv", "sync", "--project", runnerPath, "--compile-bytecode")
	if log, err := cmd.CombinedOutput(); err != nil {
		t.Log(string(log))
		t.Fatal("uv sync error:", err)
	}

	path, err := filepath.Abs(filepath.Join(runnerPath, ".venv"))
	if err != nil {
		t.Fatal("failed to get absolute path of reusable Python venv:", err)
	}

	return path
}

// DeletePythonVenv removes the Python virtual environment
// created by [CreatePythonVenv], during test-suite cleanup.
func DeletePythonVenv(t *testing.T, path string) {
	if err := os.RemoveAll(path); err != nil {
		t.Error("failed to remove reusable Python venv:", err)
	}
}
