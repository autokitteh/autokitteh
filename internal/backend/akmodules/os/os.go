package os

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("os"))

	retCtor = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("exec")))
)

func New() sdkexecutor.Executor {
	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportFunction("command", command, sdkmodule.WithArgs("cmd", "*args", "write=?", "read=?")),
		sdkmodule.ExportFunction("shell", shell, sdkmodule.WithArgs("cmd", "sh=?", "write=?", "read=?")),
		sdkmodule.ExportFunction("getenv", getenv, sdkmodule.WithArgs("name")),
	)
}

func execute(ctx context.Context, name string, args []string, write map[string]any, read []string) (sdktypes.Value, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 250 * time.Millisecond // without this, cancellations do not work properly.

	dir, err := os.MkdirTemp("", "autokitteh")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	cmd.Dir = dir

	mkPath := func(p string) (string, error) {
		if strings.Contains(p, "..") {
			return "", errors.New("write key must not contain '..'")
		}

		if filepath.IsAbs(p) {
			return "", errors.New("write key must be a relative path")
		}

		p = filepath.Clean(p)

		return filepath.Join(dir, p), nil
	}

	for k, v := range write {
		var data []byte

		if s, ok := v.(string); ok {
			data = []byte(s)
		} else if bs, ok := v.([]byte); ok {
			data = bs
		} else {
			return sdktypes.InvalidValue, errors.New("write value must be a string or bytes")
		}

		p, err := mkPath(k)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if err := os.WriteFile(p, data, 0o644); err != nil {
			return sdktypes.InvalidValue, err
		}
	}

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

	files := make(map[string]sdktypes.Value, len(read))
	for _, r := range read {
		p, err := mkPath(r)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		bs, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				files[r] = sdktypes.Nothing
			}

			return sdktypes.InvalidValue, err
		}

		files[r] = sdktypes.NewBytesValue(bs)
	}

	return sdktypes.NewStructValue(
		retCtor,
		map[string]sdktypes.Value{
			"output":   sdktypes.NewStringValue(string(out)),
			"exitcode": sdktypes.NewIntegerValue(rc),
			"files":    sdktypes.NewDictValueFromStringMap(files),
		},
	)
}

func command(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		cmd     string
		cmdArgs []string
		write   map[string]any
		read    []string
	)

	err := sdkmodule.UnpackArgs(args, kwargs, "cmd", &cmd, "*args", &cmdArgs, "write=?", &write, "read=?", &read)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return execute(ctx, cmd, cmdArgs, write, read)
}

func shell(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		cmd, sh string
		write   map[string]any
		read    []string
	)

	err := sdkmodule.UnpackArgs(args, kwargs, "cmd", &cmd, "sh=?", &sh, "write=?", &write, "read=?", &read)
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

	return execute(ctx, sh, cmdArgs, write, read)
}

func getenv(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var name string

	err := sdkmodule.UnpackArgs(args, kwargs, "name", &name)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewStringValue(os.Getenv(name)), nil
}
