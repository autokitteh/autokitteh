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
	"fmt"
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

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".txtar") {
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

func setUpTest(t *testing.T, akPath string, cfg map[string]any) string {
	setFirstUser()

	// Start the AK server.
	ctx := t.Context()
	svc, addr, err := startAKServer(t, ctx, akPath, cfg)
	if err != nil {
		t.Fatalf("start AK server error: %v", err)
	}

	// Eventual cleanup when the test is done.
	t.Cleanup(func() {
		if err := svc.Stop(t.Context()); err != nil {
			t.Log(fmt.Errorf("stop AK server: %w", err))
		}
	})

	return addr
}

func runTestSteps(t *testing.T, steps []string, akPath, akAddr string, cfg *testConfig) {
	var (
		actionIndex int
		ak          *tests.AKResult
		pendingReq  *httpRequest
		httpResp    *httpResponse
	)
	for i, step := range steps {
		// Skip empty lines and comments.
		if step == "" {
			continue
		}

		step = expandCapture(step)

		if step == "exit" {
			t.Log("exiting test")
			break
		}

		// Actions: ak, http, wait.
		if actions.MatchString(step) {
			// Before starting a new action, if there's a pending HTTP
			// request, send it first. We implement it this way to
			// support optional customizations below the action.
			if pendingReq != nil {
				resp, err := sendRequest(akAddr, *pendingReq)
				if err != nil {
					t.Errorf("line %d: %s", actionIndex+1, steps[actionIndex])
					// Fail-fast, don't run subsequent test steps.
					t.Fatalf("error: %v", err)
				}
				pendingReq = nil
				httpResp = resp
			}

			// Now start with the new action, and store its result.
			actionIndex = i
			result, err := runAction(t, akPath, akAddr, i, step, cfg)
			if err != nil {
				t.Errorf("line %d: %s", i+1, step)
				// Fail-fast, don't run subsequent test steps.
				t.Fatalf("error: %v", err)
			}

			if result != nil {
				switch v := result.(type) {
				case *tests.AKResult:
					ak = v
				case *httpRequest:
					pendingReq = v
				case string:
					t.Log(v)
				default:
					t.Errorf("line %d: %s", i+1, step)
					t.Fatalf("error: unhandled action result type: %T", v)
				}
			}

			continue
		}

		// Before running a check, if it's an HTTP check and there's a pending
		// HTTP request, send the request first. We implement it this way to
		// support optional customizations between the action and its checks.
		if httpChecks.MatchString(step) && pendingReq != nil {
			resp, err := sendRequest(akAddr, *pendingReq)
			if err != nil {
				t.Errorf("line %d: %s", actionIndex+1, steps[actionIndex])
				// Fail-fast, don't run subsequent test steps.
				t.Fatalf("error: %v", err)
			}
			pendingReq = nil
			httpResp = resp
		}

		// Checks: ak output, ak return code, http resp.
		if err := runCheck(t, step, ak, httpResp); err != nil {
			t.Errorf("line %d: %s", actionIndex+1, steps[actionIndex])
			t.Errorf("line %d: %s", i+1, step)
			// Fail-fast, don't run subsequent test steps.
			t.Fatalf("error: %v", err)
		}
	}
}
