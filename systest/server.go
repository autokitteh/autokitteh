package systest

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const (
	serverReadyTimeout = 10 * time.Second
	serverAddrFilename = "AK_SERVER_ADDR"
)

func startAKServer(ctx context.Context, combinedOutput *bytes.Buffer) {
	cmd.RootCmd.SetArgs([]string{
		"up",
		"--config", "http.addr=:0",
		"--config", "http.addr_filename=" + serverAddrFilename,
		"--mode", "test",
	})
	cmd.RootCmd.SetOut(combinedOutput)
	cmd.RootCmd.SetErr(combinedOutput)
	// This is blocking for the entire duration of the test, so it's safe to
	// ignore the error, and even to abort the goroutine with an internal panic.
	kittehs.Must0(cmd.RootCmd.ExecuteContext(ctx))
}

func waitForAKServer(t *testing.T, combinedOutput *bytes.Buffer) string {
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
		t.Fatalf("ak server output:\n%s", combinedOutput)
	}

	// Return the AK server's address, to be used by clients/tools.
	akAddr, err := os.ReadFile(serverAddrFilename)
	if err != nil {
		t.Errorf("failed to read ak server address: %v", err)
		t.Fatalf("ak server output:\n%s", combinedOutput.String())
	}
	os.Remove(serverAddrFilename)
	combinedOutput.Reset()
	return string(akAddr)
}

func checkAKServer(combinedOutput *bytes.Buffer, result chan<- time.Duration) {
	ready := []byte("ready")
	start := time.Now()
	for {
		if bytes.Contains(combinedOutput.Bytes(), ready) {
			result <- time.Since(start)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
