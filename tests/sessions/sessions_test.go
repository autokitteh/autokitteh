package sessions

import (
	"embed"
	"flag"
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

// If set, session test will run with --durable. Any *.durable.* tests will be run, and
// any *.nondurable.* tests will be skipped.
// If not set, session test will run without --durable. Any *.nondurable.* tests will be run,
// and any *.durable.* tests will be skipped.
var (
	durable = flag.Bool("durable", false, "run sessions tests in durable mode")
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

		if (strings.Contains(txtarPath, ".nondurable.") || strings.Contains(txtarPath, ".nondurable/")) && durable != nil && *durable {
			t.Skip("skipping in durable mode")
			return
		}

		if (strings.Contains(txtarPath, ".durable.") || strings.Contains(txtarPath, ".durable/")) && (durable == nil || !*durable) {
			t.Skip("skipping in nondurable mode")
			return
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
		result, err := tests.RunAKClient(t, akPath, server.Addr, "", clientTimeout, args)
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

		if durable != nil && *durable {
			args = append(args, "--durable")
		}

		result, err = tests.RunAKClient(t, akPath, server.Addr, "", clientTimeout, args)
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
