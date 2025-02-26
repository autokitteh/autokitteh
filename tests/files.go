package tests

import (
	"os"
	"path/filepath"
	"testing"
)

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
func SwitchToTempDir(t *testing.T) string {
	path := t.TempDir()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get current working directory:", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal("failed to restore working directory:", err)
		}
	})

	if err := os.Chdir(path); err != nil {
		t.Fatal("failed to switch to temporary directory:", err)
	}

	// Don't use the user's "config.yaml" file, it may violate isolation
	// by forcing tests to use shared and/or persistent resources.
	t.Setenv("XDG_CONFIG_HOME", path)
	t.Setenv("XDG_DATA_HOME", path)

	return path
}
