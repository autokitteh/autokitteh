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
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	rootDir     = "testdata/"
	stopTimeout = 3 * time.Second
)

func TestSuite(t *testing.T) {
	akPath := setUpSuite(t)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err) // Abort the entire test suite on walking errors.
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".txtar") {
			return nil // Skip directories and non-test files.
		}

		// Each .txtar file is a test-case, with potentially
		// multiple actions, checks, and embedded files.
		t.Run(strings.TrimPrefix(path, rootDir), func(t *testing.T) {
			steps := readTestFile(t, path)
			akAddr := setUpTest(t, akPath)
			runTestSteps(t, steps, akPath, akAddr)
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func setUpSuite(t *testing.T) string {
	akPath := buildClient(t)

	// https://docs.temporal.io/dev-guide/go/debugging
	t.Setenv("TEMPORAL_DEBUG", "true")

	return akPath
}

func setUpTest(t *testing.T, akPath string) string {
	// TODO: Replace "/backend/internal/temporalclient/client.go"?

	// Redirect the OS's stdout and stderr through a pipe, to
	// detect when the AK server is ready for the test to begin.
	origStdout, origStderr := os.Stdout, os.Stderr
	combinedOutput := newMutexBuffer()
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Stderr = w
	go io.Copy(combinedOutput, r) //nolint:all

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		r.Close() // End the io.Copy goroutine.
		w.Close()
	}()

	// Start the AK server, but in-process rather than as a separate
	// subprocess: to support breakpoint debugging, and measure test coverage.
	svc, err := startAKServer(context.Background(), akPath)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
		defer cancel()
		if err := svc.Stop(ctx); err != nil {
			t.Log(fmt.Errorf("fx app stop: %w", err))
		}
	})

	return svc.Addr()
}

func runTestSteps(t *testing.T, steps []string, akPath, akAddr string) {
	var (
		actionIndex int
		ak          *akResult
		pendingReq  *httpRequest
		httpResp    *httpResponse
	)
	for i, step := range steps {
		// Skip empty lines and comments.
		if step == "" {
			continue
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
			result, err := runAction(t, akPath, akAddr, step)
			if err != nil {
				t.Errorf("line %d: %s", i+1, step)
				// Fail-fast, don't run subsequent test steps.
				t.Fatalf("error: %v", err)
			}

			if result != nil {
				switch v := result.(type) {
				case *akResult:
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
		if err := runCheck(step, ak, httpResp); err != nil {
			t.Errorf("line %d: %s", actionIndex+1, steps[actionIndex])
			t.Errorf("line %d: %s", i+1, step)
			// Fail-fast, don't run subsequent test steps.
			t.Fatalf("error: %v", err)
		}
	}
}
