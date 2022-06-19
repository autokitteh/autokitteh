package internalplugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"go.autokitteh.dev/sdk/api/apivalues"
)

func run(ctx context.Context, path string, args, env []interface{}, dir string, fail bool) (*apivalues.Value, error) {
	sargs := make([]string, len(args))
	for i, a := range args {
		s, ok := a.(string)
		if !ok {
			return nil, fmt.Errorf("args must be a list of strings")
		}

		sargs[i] = s
	}

	var outbuf bytes.Buffer

	cmd := exec.CommandContext(ctx, path, sargs...)
	cmd.Dir = dir
	cmd.SysProcAttr = osSpecificSysProcAttr()
	cmd.Stdout = &outbuf
	cmd.Stderr = &outbuf

	output := func() string { return outbuf.String() }

	if env != nil {
		cmd.Env = make([]string, len(env))
		for i, e := range env {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("env must be a list of strings")
			}

			cmd.Env[i] = s
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w\n%s", err, output())
	}

	var exited bool
	defer func() {
		exited = true
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		// Wait for context to be canceled
		<-ctx.Done()

		// If the process has already exited no need to try to kill it
		if exited {
			return
		}

		// Send sigint to the process gorup and wait for some time to allow for graceful shutdown
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT); err != nil {
			if errors.Is(os.ErrProcessDone, err) {
				return
			}

			// If there was an error sending sigint just send kill
			// Using -pid will send the kill signal to process group
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			_ = cmd.Process.Kill()
		}

		// Check frequently until the process has exited
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			<-time.After(200 * time.Millisecond)
			if exited {
				return
			}
		}

		// The process hasn't exited, try to kill it again and abandon ship
		// Using -pid will send the kill signal to process group
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = cmd.Process.Kill()
	}()

	if err := cmd.Wait(); err != nil {
		if _, ee := err.(*exec.ExitError); fail || !ee {
			return nil, fmt.Errorf("%w\n%s", err, output())
		}
	}

	return apivalues.MustNewValue(apivalues.ListValue(
		[]*apivalues.Value{
			apivalues.Integer(int64(cmd.ProcessState.ExitCode())),
			apivalues.String(output()),
		},
	)), nil
}
