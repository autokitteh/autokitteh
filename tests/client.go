package tests

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type AKResult struct {
	Output     string
	ReturnCode int
}

// RunAKClient runs the AK client as a subprocess, not a goroutine,
// to ensure isolation with the server and other client executions.
// The first 2 "ak*" parameters are required, the rest are optional.
// It is assumed that the AK binary was built before running the test.
func RunAKClient(t *testing.T, akPath, akAddr, userToken string, timeout time.Duration, args []string) (*AKResult, error) {
	ctx := t.Context()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if akAddr != "" {
		args = append(args, "--config", "http.service_url=http://"+akAddr)
	}

	if userToken != "" {
		args = append(args, "--token", userToken)
	}

	cmd := exec.CommandContext(ctx, akPath, args...)
	cmd.WaitDelay = time.Second // Kill the process on timeout.
	output, err := cmd.CombinedOutput()

	r := &AKResult{
		Output:     strings.TrimSpace(string(output)),
		ReturnCode: cmd.ProcessState.ExitCode(),
	}

	// Don't report non-zero-exit-code errors as errors. Both cases end up
	// as test failures, but errors are reserved for unexpected failures.
	if ee := new(exec.ExitError); errors.As(err, &ee) {
		err = nil
	}

	return r, err
}
