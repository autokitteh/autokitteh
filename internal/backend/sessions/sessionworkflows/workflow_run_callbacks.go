package sessionworkflows

import (
	"context"
	"errors"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type runCallbacks struct {
	w    *sessionWorkflow
	wctx workflow.Context
	l    *zap.Logger

	run   func() sdkservices.Run
	print func(string)
}

func (cb runCallbacks) NewRunID() (runID sdktypes.RunID) {
	if err := workflow.SideEffect(cb.wctx, func(workflow.Context) any {
		return sdktypes.NewRunID()
	}).Get(&runID); err != nil {
		cb.l.With(zap.Error(err)).Sugar().Panicf("new run ID side effect: %v", err)
	}
	return
}

func (cb runCallbacks) Load(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
	return cb.w.load(ctx, rid, path)
}

func (cb runCallbacks) Call(callCtx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	l := cb.l

	if f := v.GetFunction(); f.HasFlag(sdktypes.ConstFunctionFlag) {
		l.Debug("const function call")
		return f.ConstValue()
	}

	isActivity := activity.IsActivity(callCtx)

	l = l.With(zap.Any("run_id", runID), zap.Bool("is_activity", isActivity), zap.Any("v", v))

	if cb.run() == nil {
		l.Debug("run not initialized")
		return sdktypes.InvalidValue, errors.New("cannot call before the run is initialized")
	}

	if xid := v.GetFunction().ExecutorID(); xid.ToRunID() == runID && cb.w.executors.GetCaller(xid) == nil {
		l.Debug("self call during initialization")

		// This happens only during initial evaluation (the first run because invoking the entrypoint function),
		// and the runtime tries to call itself in order to start an activity with its own functions.
		return sdktypes.InvalidValue, errors.New("cannot call self during initial evaluation")
	}

	if isActivity {
		l.Debug("nested activity call")
		return sdktypes.InvalidValue, errors.New("nested activities are not supported")
	}

	return cb.w.call(cb.wctx, runID, v, args, kwargs)
}

func (cb runCallbacks) Print(printCtx context.Context, runID sdktypes.RunID, text string) {
	wctx := cb.wctx
	sid := cb.w.data.Session.ID()
	svcs := cb.w.ws.svcs

	isActivity := activity.IsActivity(printCtx)

	// Trim single trailing space, but no other spaces.
	if text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}

	l := cb.l.With(zap.Any("run_id", runID), zap.Bool("is_activity", isActivity))
	l.Debug("print", zap.String("text", text))

	cb.print(text)

	var err error

	if cb.run() == nil || isActivity {
		// run == nil: Since initial run is running via temporalclient.LongRunning, we
		// cannot use the workflow context to report prints.
		//
		// TODO(ENG-1554): we need to retry here.
		err = svcs.DB.AddSessionPrint(printCtx, sid, text)
	} else {
		// TODO(ENG-1467): Currently the python runtime is calling Print from a
		//                 separate goroutine. This is a workaround to detect
		//                 that behaviour and in that case just write the
		//                 print directly into the DB since we cannot launch
		//                 an activity from a separate goroutine.
		if temporalclient.IsWorkflowContextAsGoContext(printCtx) {
			err = workflow.ExecuteActivity(wctx, addSessionPrintActivityName, sid, text).Get(wctx, nil)
		} else if !workflow.IsReplaying(wctx) {
			err = svcs.DB.AddSessionPrint(printCtx, sid, text)
		}
	}

	// We do not consider print failure as a critical error, since we don't want to hold back the
	// workflow for potential user debugging prints. Just log the error and move on. Nevertheless,
	// this is a problem to be aware of, because errors cause the loss of valuable debugging data
	// (because the workflow context is canceled).
	if err != nil {
		l.With(zap.String("text", text)).Sugar().Warnf("failed to add print session record: %v", err)
	}
}
