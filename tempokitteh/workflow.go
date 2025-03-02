package tempokitteh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	envVarsModuleName = "env"
)

var envVarsExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID(envVarsModuleName))

type tkworkflow struct {
	tk *tk

	l *zap.Logger

	sessionID sdktypes.SessionID

	executors sdkexecutor.Executors
	callSeq   uint32

	children map[string]workflow.ChildWorkflowFuture
}

func (tk *tk) workflow(wctx workflow.Context, arg any) (any, error) {
	tkw := &tkworkflow{tk: tk, sessionID: sdktypes.NewSessionID(), children: make(map[string]workflow.ChildWorkflowFuture)}

	globals, err := tkw.globals()
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	sdkArg, err := sdktypes.WrapValue(arg)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("args: %w", err)
	}

	info := workflow.GetInfo(wctx)
	name := info.WorkflowType.Name
	ep, ok := tk.entrypoints[name]
	if !ok {
		return sdktypes.InvalidValue, fmt.Errorf("entry point %q not found", name)
	}

	tkw.l = tk.l.Named(ep.Name())

	var unwRetVal any

	retVal, err := tkw.run(wctx, globals, sdkArg, ep)
	if err != nil {
		return nil, err
	} else if unwRetVal, err = vw.Unwrap(retVal); err != nil {
		err = fmt.Errorf("retval: %w", err)
	}

	if pwx := info.ParentWorkflowExecution; pwx != nil {
		dwctx, cancel := workflow.NewDisconnectedContext(wctx)
		defer cancel()

		wid := info.WorkflowExecution.ID

		sig := map[string]any{"id": wid}

		if err != nil {
			sig["err"] = err.Error()
		} else {
			sig["value"] = unwRetVal
		}

		if err := workflow.SignalExternalWorkflow(
			dwctx,
			pwx.ID,
			"",
			wid,
			sig,
		).Get(dwctx, nil); err != nil {
			tk.l.With(zap.Error(err)).Sugar().Errorf("signal parent session: %v", err)
		}
	}

	return unwRetVal, err
}

func (w *tkworkflow) globals() (map[string]sdktypes.Value, error) {
	m := kittehs.Must1(sdktypes.NewStructValue(sdktypes.NewSymbolValue(sdktypes.NewSymbol(envVarsModuleName)), nil))

	if err := w.executors.AddExecutor(
		envVarsModuleName,
		sdkexecutor.NewExecutor(
			nil,
			envVarsExecutorID,
			kittehs.ListToMap(os.Environ(), func(s string) (string, sdktypes.Value) {
				k, v, _ := strings.Cut(s, "=")
				return k, sdktypes.NewStringValue(v)
			}),
		),
	); err != nil {
		return nil, err
	}

	return map[string]sdktypes.Value{envVarsModuleName: m}, nil
}

func (w *tkworkflow) newRunID(wctx workflow.Context) (runID sdktypes.RunID, err error) {
	if err = workflow.SideEffect(wctx, func(workflow.Context) any {
		return sdktypes.NewRunID()
	}).Get(&runID); err != nil {
		panic(fmt.Errorf("new run ID side effect failed: %w", err))
	}
	return
}

func (w *tkworkflow) run(
	wctx workflow.Context,
	globals map[string]sdktypes.Value,
	arg sdktypes.Value,
	entrypoint sdktypes.CodeLocation,
) (sdktypes.Value, error) {
	runID, err := w.newRunID(wctx)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("new run id: %w", err)
	}

	cbs := newCallbacks(wctx, w)

	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	var run sdkservices.Run

	temporalclient.WithoutDeadlockDetection(
		wctx,
		func() {
			run, err = sdkruntimes.Run(
				ctx,
				sdkruntimes.RunParams{
					Runtimes:             w.tk.runtimes,
					BuildFile:            w.tk.build,
					Globals:              globals,
					RunID:                runID,
					FallthroughCallbacks: cbs,
					EntryPointPath:       entrypoint.Path(),
					SessionID:            w.sessionID,
				},
			)
		},
	)

	if err != nil {
		return sdktypes.InvalidValue, err
	}

	kittehs.Must0(w.executors.AddExecutor(fmt.Sprintf("run_%s", run.ID().Value()), run))

	var retVal sdktypes.Value

	// Run call only if the entrypoint includes a name.
	if epName := entrypoint.Name(); epName != "" {
		callValue, ok := run.Values()[epName]
		if !ok {
			// The user specified an entry point that does not exist.
			// WrapError so it will be a program error and not considered as an internal error.
			return sdktypes.InvalidValue, sdktypes.WrapError(fmt.Errorf("entry point %q not found after evaluation", epName)).ToError()
		}

		if !callValue.IsFunction() {
			// The user specified an entry point that is not a function.
			// WrapError so it will be a program error and not considered as an internal error.
			return sdktypes.InvalidValue, sdktypes.WrapError(fmt.Errorf("entry point %q is not a function", epName)).ToError()
		}

		if callValue.GetFunction().ExecutorID().ToRunID() != runID {
			return sdktypes.InvalidValue, errors.New("entry point does not belong to main run")
		}

		inputs := map[string]sdktypes.Value{
			"data":       arg,
			"session_id": sdktypes.NewStringValue(workflow.GetInfo(wctx).WorkflowExecution.ID),
		}

		if retVal, err = run.Call(ctx, callValue, nil, inputs); err != nil {
			return sdktypes.InvalidValue, err
		}
	}

	return retVal, nil
}

func (w *tkworkflow) call(ctx workflow.Context, _ sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	w.callSeq++

	l := w.l.With(zap.Uint32("seq", w.callSeq))

	w.l.Info("call", zap.String("function", v.String()))

	result, err := w.tk.calls.Call(ctx, &sessioncalls.CallParams{
		SessionID: w.sessionID,
		CallSpec:  sdktypes.NewSessionCallSpec(v, args, kwargs, w.callSeq),
		Executors: &w.executors,
	})
	if err != nil {
		l.Error("call failed", zap.Error(err))
		return sdktypes.InvalidValue, err
	}

	l.Info("call result", zap.String("result", result.String()))

	return result.ToPair()
}

func (w *tkworkflow) load(ctx context.Context, _ sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
	w.l.Info("load", zap.String("path", path))

	vs := w.executors.GetValues(path)
	if vs == nil {
		return nil, sdkerrors.ErrNotFound
	}

	return vs, nil
}
