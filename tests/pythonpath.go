package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// SetPythonPath prepends the AutoKitteh Python SDK's source
// code to the PYTHONPATH environment variable, so tests use
// that instead of the version published in PyPI.
func SetPythonPath(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get current working directory:", err)
	}

	path, err = filepath.Abs(filepath.Join(path, "..", "..", "runtimes", "pythonrt", "py-sdk"))
	if err != nil {
		t.Fatal("failed to construct Py-SDK path:", err)
	}

	t.Setenv("PYTHONPATH", path+":")
}
