package systest

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/tests/internal/svcproc"
)

const (
	serverHTTPAddrFile = "ak_server_addr.txt"
	serverReadyTimeout = 5 * time.Second
)

// Start the AK server, as an in-process goroutine rather than a separate
// subprocess (to support breakpoint debugging, and measure test coverage),
// unless the environment variable AK_SYSTEST_USE_PROC_SVC is set to "true".
func startAKServer(ctx context.Context, akPath string) (svc.Service, string, error) {
	runOpts := svc.RunOptions{Mode: "test"}
	cfg := kittehs.Must1(svc.LoadConfig("", map[string]any{
		"db.type": "sqlite",
		"db.dsn":  "file:autokitteh.sqlite", // In the test's temporary directory.

		"http.addr":          ":0",
		"http.addr_filename": serverHTTPAddrFile, // In the test's temporary directory.
	}, ""))

	// Instantiate the server, either as a subprocess or in-process.
	var (
		server svc.Service
		err    error
	)
	if subproc, _ := strconv.ParseBool(os.Getenv("AK_SYSTEST_USE_PROC_SVC")); subproc {
		server, err = svcproc.NewSvcProc(akPath, cfg, runOpts)
	} else {
		server, err = svc.New(cfg, runOpts)
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
	ready := make(chan bool, 1)
	timer := time.NewTimer(serverReadyTimeout)
	go queryReadyz(ready)

	select {
	case <-ready:
		timer.Stop()
	case <-timer.C:
		return "", fmt.Errorf("ak server not ready after %s", serverReadyTimeout)
	}

	// Success.
	addr, _ := os.ReadFile(serverHTTPAddrFile)
	return strings.TrimSpace(string(addr)), nil
}

func queryReadyz(result chan<- bool) {
	start := time.Now()
	for time.Since(start) < serverReadyTimeout {
		b, err := os.ReadFile(serverHTTPAddrFile)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		addr := strings.TrimSpace(string(b))
		resp, err := sendRequest(addr, httpRequest{method: "GET", url: "/readyz"})
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// For now, the availability of "/readyz" is sufficient,
		// no need to check the response body yet.
		if resp.resp.StatusCode == 200 {
			result <- true
			return
		}

		time.Sleep(10 * time.Millisecond)
	}
}
