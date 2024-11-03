// This closely follows https://github.com/google/starlark-go/tree/master/lib/time.
package testtools

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("testtools"))

var (
	// A soft error is a regular call error, can be interpreted as regular integration error.
	errTestSoft = errors.New("soft test error")

	// A hard eror is intended to be a simulated infrastructure error to trigger a retry.
	ErrTestHard = sdkerrors.NewRetryableErrorf("hard test error")
)

func New() sdkexecutor.Executor {
	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportFunction("freeze", freeze, sdkmodule.WithFlags(sdktypes.PureFunctionFlag, sdktypes.PrivilidgedFunctionFlag)),
		sdkmodule.ExportFunction("fail_activity", failActivity, sdkmodule.WithFlags(sdktypes.DisableAutoHeartbeat)),
		sdkmodule.ExportFunction("fail_workflow", failWorkflow, sdkmodule.WithFlags(sdktypes.PureFunctionFlag)),
	)
}

// There is no need to count retries here as by default workflows are not retried.
// If we decide in the future to do workflow retries, we'll add retries here.
func failWorkflow(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var hard bool

	if err := sdkmodule.UnpackArgs(args, kwargs, "hard?", &hard); err != nil {
		return sdktypes.InvalidValue, err
	}

	if hard {
		return sdktypes.InvalidValue, ErrTestHard
	}

	return sdktypes.InvalidValue, errTestSoft
}

func failActivity(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		n    int  // how many times to fail.
		hard bool // hard fail is simulating an infrastructure error
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "n?", &n, "hard?", &hard); err != nil {
		return sdktypes.InvalidValue, err
	}

	var last int

	// Use heartbeat to keep track of how many times we failed.
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &last); err != nil {
			return sdktypes.InvalidValue, err
		}
	}

	if last >= n {
		return sdktypes.NewIntegerValue(last), nil
	}

	activity.RecordHeartbeat(ctx, last+1)

	if hard {
		return sdktypes.InvalidValue, ErrTestHard
	}

	return sdktypes.InvalidValue, errTestSoft
}

// returns true if finished after replay, false if did not deadlock.
func freeze(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		t            time.Duration
		many         bool
		ignoreCancel bool
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "t", &t, "many?", &many, "ignore_cancel?", &ignoreCancel); err != nil {
		return sdktypes.InvalidValue, err
	}

	if !many {
		wctx := sessioncontext.GetWorkflowContext(ctx)

		replaying := workflow.IsReplaying(wctx)

		// This makes sure that the IsReplaying function above will
		// return true if if the workflow froze the last iteration.
		// If this wouldn't be here, and the workflow froze the last
		// replay, there would be nothing to replay and IsReplay would
		// return false.
		if err := workflow.Sleep(wctx, time.Millisecond); err != nil {
			return sdktypes.InvalidValue, err
		}

		// we abort only after the sleep in order to prevent nondeterminism.
		if replaying {
			return sdktypes.TrueValue, nil
		}
	}

	var done <-chan struct{}
	if !ignoreCancel {
		done = ctx.Done()
	}

	select {
	case <-done:
		return sdktypes.InvalidValue, ctx.Err()
	case <-time.After(t):
		// this intentionally not calling the workflow's sleep - we want
		// to trigger Temporal's deadlock detector if t > deadlock_timeout.
		return sdktypes.FalseValue, nil
	}
}
