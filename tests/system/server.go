package systest

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/tests/internal/svcproc"
)

const (
	serverHTTPAddrFile = "ak_server_addr.txt"
	serverReadyTimeout = 40 * time.Second
)

func writeSeedObjects(t *testing.T) (string, error) {
	for _, sobj := range seedObjects {
		t.Logf("seed object: %s", sobj)
	}

	path := filepath.Join(t.TempDir(), "autokitteh_seed_objects.json")

	bs, err := json.Marshal(kittehs.Transform(seedObjects, sdktypes.NewAnyObject))
	if err != nil {
		return "", fmt.Errorf("marshal seed objects: %w", err)
	}

	if err := os.WriteFile(path, bs, 0o644); err != nil {
		return "", fmt.Errorf("write seed objects: %w", err)
	}

	return path, nil
}

// Start the AK server, as an in-process goroutine rather than a separate
// subprocess (to support breakpoint debugging, and measure test coverage),
// unless the environment variable AK_SYSTEST_USE_PROC_SVC is set to "true".
func startAKServer(t *testing.T, ctx context.Context, akPath string, userCfg map[string]any) (svc.Service, string, error) {
	seedObjectsPath, err := writeSeedObjects(t)
	if err != nil {
		return nil, "", fmt.Errorf("write seed objects: %w", err)
	}

	runOpts := svc.RunOptions{Mode: "test"}

	cfgMap := map[string]any{
		"db.type": "sqlite",
		"db.dsn":  "file:autokitteh.sqlite", // In the test's temporary directory.

		"pprof.enable":                        "false",
		"http.addr":                           ":0",
		"http.addr_filename":                  serverHTTPAddrFile, // In the test's temporary directory.
		"authhttpmiddleware.use_default_user": "false",
		"svc.seed_objects_path":               seedObjectsPath,
	}

	maps.Copy(cfgMap, userCfg)

	// Instantiate the server, either as a subprocess or in-process.
	var server svc.Service

	if subproc, _ := strconv.ParseBool(os.Getenv("AK_SYSTEST_USE_PROC_SVC")); subproc {
		server, err = svcproc.NewSvcProc(akPath, cfgMap, runOpts)
	} else {
		server, err = svc.New(kittehs.Must1(svc.LoadConfig("", cfgMap, "")), runOpts)
	}
	if err != nil {
		return nil, "", fmt.Errorf("new AK server: %w", err)
	}

	// Start the server instance.
	if err := server.Start(ctx); err != nil {
		return nil, "", fmt.Errorf("start AK server: %w", err)
	}

	// Wait for the server's "/readyz" URL to be available.
	addr, err := waitForReadiness()
	if err != nil {
		return nil, "", fmt.Errorf("wait for AK server: %w", err)
	}

	return server, addr, nil
}

// Wait for the AK server to be ready, with a timeout,
// and return its HTTP address for client connections.
func waitForReadiness() (string, error) {
	ready := make(chan string, 1)
	timer := time.NewTimer(serverReadyTimeout)
	go queryReadyz(ready)

	select {
	case addr := <-ready:
		timer.Stop()
		return addr, nil // Success.
	case <-timer.C:
		return "", fmt.Errorf("ak server not ready after %s", serverReadyTimeout)
	}
}

func queryReadyz(result chan<- string) {
	start := time.Now()
	for time.Since(start) < serverReadyTimeout {
		time.Sleep(10 * time.Millisecond) // Short delay between checks.

		b, err := os.ReadFile(serverHTTPAddrFile)
		if err != nil {
			continue
		}

		addr := strings.TrimSpace(string(b))
		resp, err := sendRequest(addr, httpRequest{method: "GET", url: "/readyz"})
		if err != nil {
			continue
		}

		// For now, the availability of "/readyz" is sufficient,
		// no need to check the response body yet.
		if resp.resp.StatusCode == http.StatusOK {
			result <- addr
			return
		}
	}
}
