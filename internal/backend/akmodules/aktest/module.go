package aktest

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("test"))

func New() sdkexecutor.Executor {
	opts := []sdkmodule.Optfn{
		sdkmodule.ExportFunction(
			"panic_activity",
			panicActivity,
			sdkmodule.WithArgs("n?"),
		),
		sdkmodule.ExportFunction(
			"panic_workflow",
			panicWorkflow,
			sdkmodule.WithArgs("n?"),
			sdkmodule.WithFlags(
				sdktypes.PureFunctionFlag,
				sdktypes.PrivilidgedFunctionFlag,
			),
		),
		sdkmodule.ExportFunction(
			"freeze_activity",
			freezeActivity,
			sdkmodule.WithArgs("duration?", "allow_cancel?", "n?", "heartbeat?"),
			sdkmodule.WithFlags(
				sdktypes.DisableAutoHeartbeatFlag,
				sdktypes.ShortHeartbeatTimeout,
			),
		),
		sdkmodule.ExportFunction(
			"freeze_workflow",
			freezeWorkflow,
			sdkmodule.WithArgs("duration?", "allow_cancel?", "n?", "heartbeat?"),
			sdkmodule.WithFlags(
				sdktypes.PureFunctionFlag,
				sdktypes.PrivilidgedFunctionFlag,
			),
		),
	}

	return fixtures.NewBuiltinExecutor(ExecutorID, opts...)
}

func panicActivity(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var n int

	if err := sdkmodule.UnpackArgs(args, kwargs, "n?", &n); err != nil {
		return sdktypes.InvalidValue, err
	}

	if n > 0 {
		var count int

		if err := activity.GetHeartbeatDetails(ctx, &count); err == nil && count >= n {
			return sdktypes.Nothing, nil
		}

		activity.RecordHeartbeat(ctx, count+1)
	}

	sdklogger.Panic("panic_activity")

	return sdktypes.InvalidValue, nil
}

func panicWorkflow(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var n int

	if err := sdkmodule.UnpackArgs(args, kwargs, "n?", &n); err != nil {
		return sdktypes.InvalidValue, err
	}

	if n > 0 {
		wctx := sessioncontext.GetWorkflowContext(ctx)

		info := workflow.GetInfo(wctx)

		if int(info.Attempt) >= n {
			return sdktypes.Nothing, nil
		}
	}

	sdklogger.Panic("panic_activity")

	return sdktypes.InvalidValue, nil
}

func freezeActivity(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		duration    time.Duration
		allowCancel bool
		n           int
		heartbeat   sdktypes.Value
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "duration?", &duration, "allow_cancel?", &allowCancel, "n?", &n, "heartbeat?", &heartbeat); err != nil {
		return sdktypes.InvalidValue, err
	}

	var count int
	if n > 0 {
		if err := activity.GetHeartbeatDetails(ctx, &count); err == nil && count >= n {
			return sdktypes.Nothing, nil
		}

		count++

		activity.RecordHeartbeat(ctx, count)
	}

	var done <-chan struct{}
	if allowCancel {
		done = ctx.Done()
	}

	var tmo <-chan time.Time
	if duration > 0 {
		tmo = time.After(duration)
	}

	var heartbeatTmo time.Duration
	if heartbeat.IsValid() {
		var err error

		if b := heartbeat.GetBoolean(); b.IsValid() && b.Value() {
			heartbeatTmo = time.Second
		} else if heartbeatTmo, err = heartbeat.ToDuration(); err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("heartbeat: %w", err)
		}
	}

	var hearbeat <-chan time.Time
	if heartbeatTmo > 0 {
		hearbeat = time.After(heartbeatTmo)
	}

	for {
		select {
		case <-done:
			return sdktypes.InvalidValue, ctx.Err()
		case <-tmo:
			return sdktypes.Nothing, nil
		case <-hearbeat:
			activity.RecordHeartbeat(ctx, count)
			hearbeat = time.After(heartbeatTmo)
		}
	}
}

func freezeWorkflow(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		duration    time.Duration
		allowCancel bool
		n           int
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "duration?", &duration, "allow_cancel?", &allowCancel, "n?", &n); err != nil {
		return sdktypes.InvalidValue, err
	}

	if n > 0 {
		wctx := sessioncontext.GetWorkflowContext(ctx)

		if workflow.IsReplaying(wctx) {
			return sdktypes.Nothing, nil
		}
	}

	var done <-chan struct{}
	if allowCancel {
		done = ctx.Done()
	}

	var tmo <-chan time.Time
	if duration > 0 {
		tmo = time.After(duration)
	}

	select {
	case <-done:
		return sdktypes.InvalidValue, ctx.Err()
	case <-tmo:
		return sdktypes.Nothing, nil
	}
}
