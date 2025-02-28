package tempokitteh

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type callbacks struct {
	wctx workflow.Context
	w    *tkworkflow
}

func newCallbacks(wctx workflow.Context, w *tkworkflow) sdkservices.RunCallbacks {
	cbs := &callbacks{wctx: wctx, w: w}

	l := w.l

	return sdkservices.RunCallbacks{
		NewRunID: func() (sdktypes.RunID, error) { return w.newRunID(wctx) },
		Load:     w.load,
		Print: func(_ context.Context, runID sdktypes.RunID, text string) error {
			l.Info("print", zap.String("run_id", runID.String()), zap.String("text", text))
			return nil
		},
		Call: func(callCtx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			if f := v.GetFunction(); f.HasFlag(sdktypes.ConstFunctionFlag) {
				return f.ConstValue()
			}

			if activity.IsActivity(callCtx) {
				return sdktypes.InvalidValue, errors.New("nested activities are not supported")
			}

			if xid := v.GetFunction().ExecutorID(); xid.ToRunID() == runID && w.executors.GetCaller(xid) == nil {
				// This happens only during initial evaluation (the first run because invoking the entrypoint function),
				// and the runtime tries to call itself in order to start an activity with its own functions.
				return sdktypes.InvalidValue, errors.New("cannot call self during initial evaluation")
			}

			return w.call(wctx, runID, v, args, kwargs)
		},
		Now: func(nowCtx context.Context, runID sdktypes.RunID) (time.Time, error) {
			if activity.IsActivity(nowCtx) {
				return kittehs.Now().UTC(), nil
			}

			return workflow.Now(wctx).UTC(), nil
		},
		Sleep: func(sleepCtx context.Context, runID sdktypes.RunID, d time.Duration) error {
			if activity.IsActivity(sleepCtx) {
				select {
				case <-time.After(d):
					return nil
				case <-sleepCtx.Done():
					return sleepCtx.Err()
				}
			}

			return workflow.Sleep(wctx, d)
		},
		Start:      cbs.start,
		Signal:     cbs.signal,
		NextSignal: cbs.nextSignal,
	}
}

func (cb *callbacks) start(_ context.Context, rid sdktypes.RunID, loc sdktypes.CodeLocation, inputs map[string]sdktypes.Value, memo map[string]string) (string, error) {
	l := cb.w.l.Sugar().With("rid", rid, "loc", loc, "inputs", inputs, "memo", memo)

	l.Info("child workflow start requested")

	args, err := kittehs.TransformMapValuesError(inputs, vw.Unwrap)
	if err != nil {
		return "", fmt.Errorf("inputs: %w", err)
	}

	sid := uuid.New().String()

	f := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(
			cb.wctx,
			workflow.ChildWorkflowOptions{
				WorkflowID:        sid,
				ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
				Memo:              kittehs.TransformMapValues(memo, func(s string) any { return s }),
			},
		),
		loc.Name(),
		args,
	)

	var r workflow.Execution
	if err := f.GetChildWorkflowExecution().Get(cb.wctx, &r); err != nil {
		return "", fmt.Errorf("child workflow execution: %w", err)
	}

	l.With("workflow_id", r.ID, "run_id", r.RunID, "memo", memo).Infof("initiated child session workflow %v", r.ID)

	cb.w.children[sid] = f

	return sid, nil
}

func (cb *callbacks) signal(_ context.Context, _ sdktypes.RunID, dst, name string, v sdktypes.Value) error {
	if !v.IsValid() {
		v = sdktypes.Nothing
	}

	signal, err := vw.Unwrap(v)
	if err != nil {
		return fmt.Errorf("value: %w", err)
	}

	var f workflow.Future

	if childFuture, ok := cb.w.children[dst]; ok {
		f = childFuture.SignalChildWorkflow(cb.wctx, name, signal)
	} else {
		f = workflow.SignalExternalWorkflow(cb.wctx, dst, "", name, signal)
	}

	if err := f.Get(cb.wctx, nil); err != nil {
		return err
	}

	return nil
}

func (cb *callbacks) nextSignal(_ context.Context, _ sdktypes.RunID, names []string, timeout time.Duration) (*sdkservices.RunSignal, error) {
	if len(names) == 0 {
		return nil, nil
	}

	for i, name := range names {
		if strings.HasPrefix(name, sdktypes.SessionIDKind+"_") {
			sid, err := sdktypes.ParseSessionID(name)
			if err != nil {
				return nil, sdkerrors.NewInvalidArgumentError("invalid session id %q: %w", name, err)
			}

			names[i] = sid.String()
		} else {
			names[i] = name
		}
	}

	selector := workflow.NewSelector(cb.wctx)

	if timeout != 0 {
		selector.AddFuture(workflow.NewTimer(cb.wctx, timeout), func(workflow.Future) {})
	}

	var signal *sdkservices.RunSignal

	for _, name := range names {
		selector.AddReceive(workflow.GetSignalChannel(cb.wctx, name), func(c workflow.ReceiveChannel, _ bool) {
			var v any

			if !c.ReceiveAsync(&v) {
				cb.w.l.Warn("next_signal: expected but not received", zap.String("name", name))
			}

			payload, err := vw.Wrap(v)
			if err != nil {
				cb.w.l.Error("next_signal result:", zap.Error(err))
				return
			}

			signal = &sdkservices.RunSignal{
				Name:    name,
				Payload: payload,
			}
		})
	}

	// Select doesn't respond to cancellations unless we add a receive on the context done channel.
	var cancelled bool
	selector.AddReceive(cb.wctx.Done(), func(c workflow.ReceiveChannel, _ bool) { cancelled = true })

	selector.Select(cb.wctx)

	if cancelled {
		return nil, cb.wctx.Err()
	}

	if signal == nil {
		return nil, nil
	}

	if childFuture, ok := cb.w.children[signal.Name]; ok {
		// If we don't wait for the workflow to end, for some reason Temporal
		// have an "unknown command" meltdown and complains about non-determinism.
		_ = childFuture.Get(cb.wctx, nil)
	}

	if !signal.Payload.IsValid() {
		signal.Payload = sdktypes.Nothing
	}

	return signal, nil
}
