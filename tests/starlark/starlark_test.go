package starlark

import (
	"embed"
	"io/fs"
	"strings"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/tests"
)

const (
	clientTimeout = 15 * time.Second
)

//go:embed *
var testFiles embed.FS

func TestStarlark(t *testing.T) {
	akPath := tests.AKPath(t)

	err := fs.WalkDir(testFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".txtar") {
			return nil // Skip directories and non-test files.
		}

		runTest(t, akPath, path)
		return nil
	})
	if err != nil {
		t.Fatal("walk error:", err)
	}
}

func runTest(t *testing.T, akPath, txtarPath string) {
	t.Run(txtarPath, func(t *testing.T) {
		args := []string{"runtime", "test", "--local", "--quiet", txtarPath}
		result, err := tests.RunAKClient(t, akPath, "", "", clientTimeout, args)
		if err != nil {
			t.Fatal("runtime test error:", err)
		}
		if result.ReturnCode != 0 {
			result.Output = strings.TrimPrefix(result.Output, "Error: ")
			t.Fatalf("runtime test failed: return code = %d\n%s", result.ReturnCode, result.Output)
		}
	})
}
