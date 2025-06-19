//go:build enterprise
// +build enterprise

/*
Package test runs end-to-end "black-box" system tests on for the
autokitteh CLI tool, functioning both as a server and as a client.

It can also control other tools, dependencies, and in-memory fixtures
(e.g. Temporal, databases, caches, and HTTP webhooks).

Test cases are defined as [txtar] files in the [testdata] directory
tree. Their structure and scripting language is defined [here].

Other than local and CI/CD testing, this may be used for benchmarking,
profiling, and load/stress testing.

[txtar]: https://pkg.go.dev/golang.org/x/tools/txtar
[testdata]: https://github.com/autokitteh/autokitteh/tree/main/systest/testdata
[here]: https://github.com/autokitteh/autokitteh/tree/main/systest/README.md
*/
package systest

import (
	"embed"
	"io/fs"
	"strings"
	"testing"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/tests"
)

//go:embed *
var testFiles embed.FS

func TestSystem(t *testing.T) {
	akPath, venvPath := setUpSuite(t)

	testCases := make(map[string]*testFile)
	var exclusives []string

	// Each .txtar file is a test-case, with potentially
	// multiple actions, checks, and embedded files.
	err := fs.WalkDir(testFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".ee.txtar") {
			return nil // Skip directories and non-test files.
		}

		f, err := readTestFile(t, testFiles, path)
		if err != nil {
			return err
		}

		path = strings.TrimPrefix(path, "testdata/")
		testCases[path] = f

		// Same as the "-run" flag in "go test", but easier to use.
		if f.config.Exclusive {
			exclusives = append(exclusives, path)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Same as the "-run" flag in "go test", but easier to use.
	filter := func(string) bool { return true }
	if len(exclusives) > 0 {
		filter = kittehs.ContainedIn(exclusives...)
	}

	for path, test := range testCases {
		t.Run(path, func(t *testing.T) {
			if !filter(path) {
				t.Skip("skipping")
			}

			tests.SwitchToTempDir(t, venvPath) // For test isolation.
			akAddr := setUpTest(t, akPath, test.config.Server)

			writeEmbeddedFiles(t, test.a.Files)

			runTestSteps(t, test.steps, akPath, akAddr, &test.config)
		})
	}
}
