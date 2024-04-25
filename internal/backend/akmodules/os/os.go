package os

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("os"))

func New() sdkexecutor.Executor {
	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportFunction("command", command, sdkmodule.WithArgs("cmd", "*args")),
		sdkmodule.ExportFunction("shell", shell, sdkmodule.WithArgs("cmd", "sh?")),
	)
}

func execute(ctx context.Context, name string, args ...string) (sdktypes.Value, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 250 * time.Millisecond // without this, cancellations do not work properly.
	out, err := cmd.CombinedOutput()

	var rc int

	var exit *exec.ExitError
	if errors.As(err, &exit) {
		rc = exit.ExitCode()

		if rc < 0 && (errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled)) {
			// rc could not be determined and the context was canceled or timed out.
			err = ctx.Err()
		} else {
			// rc could be determined or no error context occured.
			err = nil
		}
	}

	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return kittehs.Must1(sdktypes.NewListValue([]sdktypes.Value{
		sdktypes.NewStringValue(string(out)),
		sdktypes.NewIntegerValue(int64(rc)),
	})), nil
}

func command(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		cmd     string
		cmdArgs []string
	)

	err := sdkmodule.UnpackArgs(args, kwargs, "cmd", &cmd, "*args", &cmdArgs)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return execute(ctx, cmd, cmdArgs...)
}

func shell(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var cmd, sh string

	err := sdkmodule.UnpackArgs(args, kwargs, "cmd", &cmd, "sh?", &sh)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	var cmdArgs []string

	if runtime.GOOS == "windows" {
		if sh == "" {
			sh = "cmd"
		}

		cmdArgs = []string{"/c"}
	} else {
		if sh == "" {
			sh = "/bin/sh"
		}

		cmdArgs = []string{"-c"}
	}

	cmdArgs = append(cmdArgs, cmd)

	return execute(ctx, sh, cmdArgs...)
}
