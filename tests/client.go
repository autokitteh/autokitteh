package tests

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

type AKResult struct {
	Output     string
	ReturnCode int
}

// RunAKClient runs the AK client as a subprocess, not a goroutine,
// to ensure isolation with the server and other client executions.
func RunAKClient(akPath, akAddr, userToken string, timeout time.Duration, args []string) (*AKResult, error) {
	ctx := context.Background()
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

	if ee := new(exec.ExitError); errors.As(err, &ee) {
		err = nil
	}

	return r, err
}
