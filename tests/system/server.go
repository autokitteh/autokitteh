package systest

import (
	"bytes"
	"context"
	"regexp"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd"
)

const (
	serverReadyTimeout = 10 * time.Second
)

// Start the AK server, but in a goroutine rather than as a separate
// subprocess: to support breakpoint debugging, and measure test coverage.
func startAKServer(ctx context.Context) {
	cmd.RootCmd.SetArgs([]string{
		"up",
		"--config", "db.dsn=sqlite:test.sqlite",
		"--config", "http.addr=:0",
		"--mode", "test",
	})

	// We don't care about execution errors here, the test will check this.
	cmd.RootCmd.ExecuteContext(ctx) //nolint:errcheck
}

func waitForAKServer(t *testing.T, combinedOutput *mutexBuffer) string {
	ready := make(chan time.Duration, 1)
	timer := time.NewTimer(serverReadyTimeout)
	go checkAKServer(combinedOutput, ready)

	// Wait for the AK server to be ready, up to the given timeout.
	select {
	case duration := <-ready:
		t.Logf("ak server ready after %s", duration.Round(time.Millisecond))
		timer.Stop()
	case <-timer.C:
		t.Errorf("ak server not ready after %s", serverReadyTimeout)
		t.Fatalf("ak server combined output: %s", combinedOutput.String())
	}

	// Return the AK server's address, to be used by clients/tools.
	re := regexp.MustCompile(`gRPC/HTTP:\s*(.*:\d+)`)
	addr := re.FindStringSubmatch(combinedOutput.String())
	if addr == nil {
		t.Error("ak server address not found")
		t.Fatalf("ak server combined output: %s", combinedOutput.String())
	}
	return addr[1]
}

func checkAKServer(combinedOutput *mutexBuffer, result chan<- time.Duration) {
	ready := []byte("ready")
	start := time.Now()
	for {
		if bytes.Contains(combinedOutput.Bytes(), ready) {
			result <- time.Since(start)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
