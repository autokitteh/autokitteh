package sessions

import (
	"embed"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/tests"
)

const (
	clientTimeout = 30 * time.Second
)

//go:embed *
var testFiles embed.FS

func TestSessions(t *testing.T) {
	akPath, venvPath := setUpSuite(t)

	err := fs.WalkDir(testFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".txtar") {
			return nil // Skip directories and non-test files.
		}

		runTest(t, akPath, venvPath, path)
		return nil
	})
	if err != nil {
		t.Fatal("walk error:", err)
	}
}

func setUpSuite(t *testing.T) (akPath, venvPath string) {
	// https://docs.temporal.io/dev-guide/go/debugging
	t.Setenv("TEMPORAL_DEBUG", "true")

	akPath = tests.AKPath(t)

	tests.SetPythonPath(t)
	venvPath = tests.CreatePythonVenv(t)
	t.Cleanup(func() {
		tests.DeletePythonVenv(t, venvPath)
	})

	return
}

func runTest(t *testing.T, akPath, venvPath, txtarPath string) {
	t.Run(txtarPath, func(t *testing.T) {
		absPath, err := filepath.Abs(txtarPath)
		if err != nil {
			t.Fatalf("failed to convert %q to absolute path: %v", txtarPath, err)
		}

		// Start AK server.
		tests.SwitchToTempDir(t, venvPath) // For test isolation.

		server, err := tests.StartAKServer(akPath, "dev")
		t.Cleanup(server.Stop)
		if err != nil {
			server.PrintLog(t)
			t.Fatal(err)
		}

		// Create project.
		projName := fmt.Sprintf("test_%d", rand.Uint32())
		args := []string{"project", "create", "--name", projName}
		result, err := tests.RunAKClient(akPath, server.Addr, "", clientTimeout, args)
		if err != nil {
			server.PrintLog(t)
			t.Fatal("project creation error:", err)
		}
		if result.ReturnCode != 0 {
			server.PrintLog(t)
			t.Fatalf("project %q creation failed: return code = %d\n%s", projName, result.ReturnCode, result.Output)
		}

		// Run session test.
		args = []string{"session", "test", absPath, "--project", projName}
		result, err = tests.RunAKClient(akPath, server.Addr, "", clientTimeout, args)
		if err != nil {
			server.PrintLog(t)
			t.Fatal("session test error:", err)
		}
		if result.ReturnCode != 0 {
			server.PrintLog(t)
			t.Fatalf("session test failed: return code = %d\n%s", result.ReturnCode, result.Output)
		}
	})
}
