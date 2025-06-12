package systest

import (
	"context"
	"encoding/csv"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"go.autokitteh.dev/autokitteh/tests"
)

func splitToArgs(cmdArgs string) []string {
	cmdArgs = strings.TrimSpace(cmdArgs)
	r := csv.NewReader(strings.NewReader(cmdArgs))
	r.Comma = ' '       // space
	r.LazyQuotes = true // allow quotes to appear in string
	fields, _ := r.Read()
	return fields
}

func runAction(t *testing.T, akPath, akAddr string, i int, step string, cfg *testConfig) (any, error) {
	t.Logf("*** ACTION: line %d: %q", i+1, step)
	match := actions.FindStringSubmatch(step)
	trimmedArgs := strings.TrimSpace(strings.TrimPrefix(step, match[1]))
	switch match[1] {
	case "user":
		return nil, setUser(trimmedArgs)
	case "setenv":
		return nil, setEnv(trimmedArgs)
	case "ak":
		args := make([]string, len(cfg.AK.ExtraArgs))
		copy(args, cfg.AK.ExtraArgs)
		args = append(args, splitToArgs(match[3])...)
		return tests.RunAKClient(t, akPath, akAddr, token, 0, args)
	case "http get", "http post":
		method := strings.ToUpper(match[2])
		url, body, _ := strings.Cut(match[3], " ")
		return &httpRequest{method: method, url: url, body: body}, nil
	case "wait":
		return waitForSession(t, akPath, akAddr, step)
	case "exec":
		return execCommand(t.Context(), trimmedArgs)
	default:
		return nil, errors.New("unhandled action")
	}
}

func execCommand(ctx context.Context, cmdline string) (*tests.AKResult, error) {
	cmdline = strings.TrimSpace(cmdline)
	if len(cmdline) == 0 {
		return nil, errors.New("empty command")
	}

	args := splitToArgs(cmdline)
	if len(args) == 0 {
		return nil, errors.New("no command provided")
	}

	if len(args) == 0 {
		return nil, errors.New("no command name provided")
	}

	name := args[0]
	args = args[1:]

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = time.Second // Kill the process on timeout.

	out, err := cmd.CombinedOutput()

	// Don't report non-zero-exit-code errors as errors. Both cases end up
	// as test failures, but errors are reserved for unexpected failures.
	if ee := new(exec.ExitError); errors.As(err, &ee) {
		err = nil
	}

	return &tests.AKResult{
		Output:     strings.TrimSpace(string(out)),
		ReturnCode: cmd.ProcessState.ExitCode(),
	}, err
}
