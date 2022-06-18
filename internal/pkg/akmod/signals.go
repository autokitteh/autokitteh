package akmod

import (
	"context"
	"fmt"
	"golang.org/x/exp/slices"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/workflow"

	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"

	"github.com/autokitteh/L"
)

type signals struct {
	l L.L
}

func (s *signals) asStruct(funcToValue pluginimpl.FuncToValueFunc) apivalues.StructValue {
	return apivalues.StructValue{
		Ctor: apivalues.Symbol("signals"),
		Fields: map[string]*apivalues.Value{
			"send": funcToValue("send", s.send, pluginimpl.WithFlags("session")),
			"wait": funcToValue("wait", s.wait, pluginimpl.WithFlags("session")),
		},
	}
}

func (s *signals) send(
	cctx context.Context,
	_ string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	_ pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	session := getSessionContext(cctx)
	ctx, l := session.Context, session.L

	var (
		v           *apivalues.Value
		name, dstid string
	)

	if err := pluginimpl.UnpackArgs(
		args,
		kwargs,
		"dst", &dstid,
		"name", &name,
		"value", &v,
	); err != nil {
		return nil, err
	}

	dstWorkflowID := events.GetIngestProjectEventWorkflowID(apievent.EventID(dstid), session.ProjectID)

	l.Debug("sending signal", "dst", dstid, "dst_wid", dstWorkflowID, "name", name, "value", v)

	fut := workflow.ExecuteLocalActivity(
		ctx,
		func(ctx context.Context, wid string, data interface{}) error {
			if err := session.Temporal.SignalWorkflow(ctx, wid, "", SessionEventSignalName, data); err != nil {
				if _, ok := err.(*serviceerror.NotFound); ok { // for some reason errors.Is doesn't work well here.
					l.Debug("workflow not found, might have already finished")
					return nil
				}

				return L.Error(l, "signal error", "err", err)
			}

			return nil
		},
		dstWorkflowID,
		NewSyntheticEvent(session.Event, name, v),
	)

	if err := fut.Get(ctx, nil); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *signals) wait(
	cctx context.Context,
	_ string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	_ pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	session := getSessionContext(cctx)
	ctx, l := session.Context, session.L

	var (
		tmo   apivalues.DurationValue
		names = make([]string, len(args))
	)

	for i, arg := range args {
		names[i] = arg.String()
	}

	if err := pluginimpl.UnpackArgs(
		nil,
		kwargs,
		"timeout?", &tmo,
	); err != nil {
		return nil, err
	}

	if len(names) == 0 && tmo == 0 {
		return nil, fmt.Errorf("no names or tmo specified")
	}

	l.Debug("waiting for signal", "names", names)

	if err := session.UpdateState(apievent.NewWaitingProjectEventState(names, session.RunSummary)); err != nil {
		return nil, fmt.Errorf("set waiting: %w", err)
	}

	var (
		sig       *sessionEventSignal
		tmoFuture workflow.Future
	)

	if tmo != 0 {
		tmoFuture = workflow.NewTimer(ctx, time.Duration(tmo))
	}

	// flush all previously sent signals.
	for session.SignalChannel.ReceiveAsync(&sig) {
	}

	// loop until a timeout occurs or a relevant signal is received.
	for {
		sel := workflow.NewSelector(ctx)

		sel.AddReceive(
			session.SignalChannel,
			func(c workflow.ReceiveChannel, _ bool) {
				c.Receive(ctx, &sig)
				l.Debug("received signal", "sig", sig)
			},
		)

		if tmoFuture != nil {
			sel.AddFuture(
				tmoFuture,
				func(f workflow.Future) {
					l.Debug("timed out")
				},
			)
		}

		sel.Select(ctx)

		l := l.With("sig", sig)

		l.Debug("select returned")

		if sig == nil {
			l.Debug("timed out")
			break
		}

		if slices.Contains(names, sig.Name) {
			l.Debug("relevant signal")
			break
		}

		l.Debug("irrelevant signal")

		sig = nil
	}

	if err := session.UpdateState(apievent.NewRunningProjectEventState()); err != nil {
		return nil, fmt.Errorf("set running: %w", err)
	}

	if sig == nil {
		return apivalues.None, nil
	}

	st := sig.Event.AsStructValue()
	st.Fields["name"] = apivalues.String(sig.Name)

	if sig.Event.EventSourceID() == syntheticEventSourceID && sig.Event.Type() == syntheticEventType {
		st.Fields["value"] = sig.Event.Data()["value"]
	}

	return apivalues.MustNewValue(st), nil
}

func (s *signals) asValue(funcToValue pluginimpl.FuncToValueFunc) *apivalues.Value {
	return apivalues.MustNewValue(s.asStruct(funcToValue))
}
