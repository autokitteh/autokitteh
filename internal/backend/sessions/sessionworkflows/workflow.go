package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	testtoolsmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/testtools"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/webhookssvc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	envVarsModuleName     = "env"
	integrationPathPrefix = "@"
)

var errForbiddenInActivity = fmt.Errorf("%w: this operation is not allowed in an activity", sdkerrors.ErrFailedPrecondition)

var envVarsExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID(envVarsModuleName))

type sessionWorkflow struct {
	l                   *zap.Logger
	ws                  *workflows
	workflowExecutionID string

	data sessiondata.Data

	// All the members below must be built deterministically by the workflow.
	// They are not persisted in the database.

	executors sdkexecutor.Executors
	globals   map[string]sdktypes.Value

	callSeq uint32

	lastReadEventSeqForSignal map[uuid.UUID]uint64 // map signals to last read event seq num.
}

type connInfo struct {
	Config          map[string]string `json:"config"`
	IntegrationName string            `json:"integration_name"`
}

func runWorkflow(
	wctx workflow.Context,
	l *zap.Logger,
	ws *workflows,
	params sessionWorkflowParams,
) (prints []sdkservices.SessionPrint, rv sdktypes.Value, err error) {
	w := &sessionWorkflow{
		l:                         l,
		data:                      params.Data,
		ws:                        ws,
		lastReadEventSeqForSignal: make(map[uuid.UUID]uint64),
		workflowExecutionID:       workflow.GetInfo(wctx).WorkflowExecution.ID,
	}

	var cinfos map[string]connInfo

	if cinfos, err = w.initConnections(wctx); err != nil {
		return
	}

	if err = w.initEnvModule(cinfos); err != nil {
		return
	}

	if w.globals, err = w.initGlobalModules(wctx); err != nil {
		return
	}

	prints, rv, err = w.run(wctx, l)

	// context might have been canceled, create a disconnected one.
	wctx, cancel := workflow.NewDisconnectedContext(wctx)
	w.cleanupSignals(wctx)
	cancel()

	return
}

func (w *sessionWorkflow) cleanupSignals(ctx workflow.Context) {
	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)
	for signalID := range w.lastReadEventSeqForSignal {
		if err := w.ws.svcs.DB.RemoveSignal(goCtx, signalID); err != nil {
			// No need to any handling in case of an error, it won't be used again at
			// most we would have db garbage we can clear up later with background jobs.
			w.l.Sugar().With("signalID", signalID, "err", err).Warnf("failed removing signal %v, err: %v", signalID, err)
		}
	}
}

func (w *sessionWorkflow) updateState(wctx workflow.Context, state sdktypes.SessionState) error {
	return w.ws.updateSessionState(wctx, w.data.Session.ID(), state)
}

func (w *sessionWorkflow) loadIntegrationConnections(path string) (map[string]sdktypes.Value, error) {
	// Since the load callback does not inform us which member exactly from the integration it wishes
	// to load, we must return all relevant connections.

	prefix := integrationModulePrefix(path)
	vs := w.executors.ValuesForScopePrefix(prefix)

	return kittehs.TransformMapError(
		vs,
		func(n string, vs map[string]sdktypes.Value) (string, sdktypes.Value, error) {
			n = strings.TrimPrefix(n, prefix)
			sym, err := sdktypes.ParseSymbol(n)
			if err != nil {
				return "", sdktypes.InvalidValue, fmt.Errorf("invalid symbol %q: %w", n, err)
			}

			return n, kittehs.Must1(sdktypes.NewStructValue(sdktypes.NewSymbolValue(sym), vs)), nil
		},
	)
}

func (w *sessionWorkflow) load(ctx context.Context, _ sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
	if strings.HasPrefix(path, integrationPathPrefix) {
		return w.loadIntegrationConnections(path[1:])
	}

	vs := w.executors.GetValues(path)
	if vs == nil {
		return nil, sdkerrors.ErrNotFound
	}

	return vs, nil
}

func (w *sessionWorkflow) outcome(wctx workflow.Context) func(ctx context.Context, runID sdktypes.RunID, v sdktypes.Value) error {
	return func(ctx context.Context, runID sdktypes.RunID, v sdktypes.Value) error {
		ctx, span := w.startCallbackSpan(ctx, "http_response")
		defer span.End()

		isActivity := activity.IsActivity(ctx)

		w.l.Debug("http_response", zap.Any("run_id", runID), zap.Bool("is_activity", isActivity))

		if isActivity {
			return w.ws.outcomeActivity(ctx, w.data.Session.ID(), v)
		} else {
			return workflow.ExecuteActivity(wctx, outcomeActivityName, w.data.Session.ID(), v).Get(wctx, nil)
		}
	}
}

func (w *sessionWorkflow) call(ctx workflow.Context, _ sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	w.callSeq++

	result, err := w.ws.calls.Call(ctx, &sessioncalls.CallParams{
		SessionID: w.data.Session.ID(),
		CallSpec:  sdktypes.NewSessionCallSpec(v, args, kwargs, w.callSeq),
		Executors: &w.executors,
	})
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return result.ToPair()
}

func (w *sessionWorkflow) initEnvModule(cinfos map[string]connInfo) error {
	vs := kittehs.ListToMap(w.data.Vars, func(v sdktypes.Var) (string, sdktypes.Value) {
		return v.Name().String(), sdktypes.NewStringValue(v.Value())
	})

	for _, conn := range w.data.Connections {
		name := conn.Name().String()
		maps.Copy(vs, kittehs.TransformMap(cinfos[name].Config, func(k, v string) (string, sdktypes.Value) {
			return fmt.Sprintf("%s__%s", name, k), sdktypes.NewStringValue(v)
		}))
		vs[name+"__connection_id"] = sdktypes.NewStringValue(conn.ID().String())

	}

	for _, t := range w.data.Triggers {
		if t.SourceType() != sdktypes.TriggerSourceTypeWebhook {
			continue
		}

		webhookURL, err := webhookssvc.WebhookSlugToAddress(t.WebhookSlug())
		if err != nil {
			w.l.Error("failed to get webhook address", zap.Any("trigger", t.Name()), zap.Error(err))
			continue
		}

		name := t.Name().String()
		vs[name+"__webhook_url"] = sdktypes.NewStringValue(webhookURL)
	}

	mod := sdkexecutor.NewExecutor(
		nil, // no calls will be ever made to env.
		envVarsExecutorID,
		vs,
	)

	kittehs.Must0(w.executors.AddExecutor(envVarsModuleName, mod))

	return nil
}

func integrationModulePrefix(name string) string { return fmt.Sprintf("__%v__", name) }

func (w *sessionWorkflow) initConnections(wctx workflow.Context) (map[string]connInfo, error) {
	// In theory, all this code is reaching external systems for integrations, but since
	// all the integrations are currently bundled with the AK binary, the operations
	// are instantaneous. No need to use activities and such right now.

	cinfos := make(map[string]connInfo, len(w.data.Connections))

	goCtx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	for _, conn := range w.data.Connections {
		name, cid, iid := conn.Name().String(), conn.ID(), conn.IntegrationID()

		intg, err := w.ws.svcs.Integrations.Attach(goCtx, iid)
		if err != nil {
			return nil, fmt.Errorf("attach to integration %q: %w", iid, err)
		}

		if intg == nil {
			return nil, fmt.Errorf("integration %q not found", iid)
		}

		uniqueName := intg.Get().UniqueName().String()

		// In modules, we register the connection prefixed with its integration name.
		// This allows us to query all connections for a given integration in the load callback.
		scope := integrationModulePrefix(uniqueName) + name
		if w.executors.GetValues(scope) != nil {
			return nil, fmt.Errorf("conflicting connection %q", name)
		}

		if xid := sdktypes.NewExecutorID(iid); w.executors.GetCaller(xid) == nil {
			kittehs.Must0(w.executors.AddCaller(xid, intg))
		}

		cinfo := connInfo{
			IntegrationName: uniqueName,
		}

		// mod's executor id is the integration id.
		var vs map[string]sdktypes.Value
		vs, cinfo.Config, err = intg.Configure(goCtx, cid)
		if err != nil {
			return nil, fmt.Errorf("connect to integration %q: %w", iid, err)
		}

		if vs == nil {
			vs = make(map[string]sdktypes.Value)
		}

		const infoVarName = "_ak_info"
		if !vs[infoVarName].IsValid() {
			if vs[infoVarName], err = sdktypes.WrapValue(cinfo); err != nil {
				return nil, fmt.Errorf("wrap conn info %q: %w", name, err)
			}
		}

		cinfos[name] = cinfo

		if err := w.executors.AddValues(scope, vs); err != nil {
			return nil, err
		}
	}

	return cinfos, nil
}

func (w *sessionWorkflow) initGlobalModules(wctx workflow.Context) (map[string]sdktypes.Value, error) {
	if !w.ws.cfg.Test {
		return nil, nil
	}

	const name = "testtools"

	tt := kittehs.Must1(sdktypes.NewStructValue(sdktypes.NewSymbolValue(sdktypes.NewSymbol(name)), nil))

	if err := w.executors.AddExecutor(name, testtoolsmodule.New(wctx)); err != nil {
		return nil, err
	}

	return map[string]sdktypes.Value{name: tt}, nil
}

func (w *sessionWorkflow) run(wctx workflow.Context, l *zap.Logger) (_ []sdkservices.SessionPrint, retVal sdktypes.Value, _ error) {
	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	startTrace := telemetry.T().Start

	ctx, workflowSpan := startTrace(ctx, "sessionWorkflow.run")
	defer workflowSpan.End()

	session := w.data.Session

	workflowSpan.SetAttributes(attribute.String("session_id", session.ID().String()))

	newRunID := func() (runID sdktypes.RunID, err error) {
		if err = workflow.SideEffect(wctx, func(workflow.Context) any {
			return sdktypes.NewRunID()
		}).Get(&runID); err != nil {
			l.With(zap.Error(err)).Sugar().Panicf("new run ID side effect: %v", err)
		}
		return
	}

	printer := w.newPrinter()

	var run sdkservices.Run

	cbs := sdkservices.RunCallbacks{
		NewRunID: newRunID,
		Load:     w.load,
		Call: func(callCtx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			callCtx, span := w.startCallbackSpan(callCtx, "call")
			defer span.End()

			span.SetAttributes(attribute.String("function_name", v.GetFunction().Name().String()))

			if f := v.GetFunction(); f.HasFlag(sdktypes.ConstFunctionFlag) {
				l.Debug("const function call")
				return f.ConstValue()
			}

			isActivity := activity.IsActivity(callCtx)

			l = l.With(zap.Any("run_id", runID), zap.Bool("is_activity", isActivity), zap.Any("v", v))

			if run == nil {
				l.Debug("run not initialized")
				return sdktypes.InvalidValue, errors.New("cannot call before the run is initialized")
			}

			if xid := v.GetFunction().ExecutorID(); xid.ToRunID() == runID && w.executors.GetCaller(xid) == nil {
				l.Debug("self call during initialization")

				// This happens only during initial evaluation (the first run because invoking the entrypoint function),
				// and the runtime tries to call itself in order to start an activity with its own functions.
				return sdktypes.InvalidValue, errors.New("cannot call self during initial evaluation")
			}

			if !session.IsDurable() {
				l.Error("call in non-durable session")
				return sdktypes.InvalidValue, errors.New("calls are not supported in non-durable sessions")
			}

			if isActivity {
				l.Debug("nested activity call")
				return sdktypes.InvalidValue, errors.New("nested activities are not supported")
			}

			return w.call(wctx, runID, v, args, kwargs)
		},
		Outcome: w.outcome(wctx),
		Print: func(printCtx context.Context, runID sdktypes.RunID, text string) error {
			ctx, span := w.startCallbackSpan(printCtx, "print")
			defer span.End()

			span.SetAttributes(attribute.String("text", text))

			isActivity := activity.IsActivity(ctx)

			l.Debug("print", zap.Any("run_id", runID), zap.Bool("is_activity", isActivity), zap.String("text", text))

			print := sdkservices.SessionPrint{
				Value: sdktypes.NewStringValue(strings.TrimSuffix(text, "\n")),
			}

			if isActivity {
				print.Timestamp = kittehs.Now()
				printer.Print(print, w.callSeq, false)
			} else {
				print.Timestamp = workflow.Now(wctx)
				printer.Print(print, 0, workflow.IsReplaying(wctx))
			}

			return nil
		},
		Now: func(nowCtx context.Context, runID sdktypes.RunID) (time.Time, error) {
			ctx, span := w.startCallbackSpan(nowCtx, "now")
			defer span.End()

			if activity.IsActivity(ctx) {
				return kittehs.Now().UTC(), nil
			}

			return workflow.Now(wctx).UTC(), nil
		},
		Sleep: func(sleepCtx context.Context, runID sdktypes.RunID, d time.Duration) error {
			ctx, span := w.startCallbackSpan(sleepCtx, "sleep")
			defer span.End()

			span.SetAttributes(attribute.Int64("d", int64(d)))

			if activity.IsActivity(ctx) {
				select {
				case <-time.After(d):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return workflow.Sleep(wctx, d)
		},
		Start:              w.start(wctx),
		Subscribe:          w.subscribe(wctx),
		Unsubscribe:        w.unsubscribe(wctx),
		NextEvent:          w.nextEvent(wctx),
		IsDeploymentActive: w.isDeploymentActive(wctx),
		Signal:             w.signal(wctx),
		NextSignal:         w.nextSignal(wctx),
		ListStoreValues:    w.listStoreValues(wctx),
		MutateStoreValue:   w.mutateStoreValue(wctx),
	}

	runID, err := newRunID()
	if err != nil {
		return nil, sdktypes.InvalidValue, fmt.Errorf("new run id: %w", err)
	}

	if err := w.updateState(wctx, sdktypes.NewSessionStateRunning(runID, sdktypes.InvalidValue)); err != nil {
		return nil, sdktypes.InvalidValue, err
	}

	entryPoint := session.EntryPoint()

	printer.Start()

	initialRunCtx, initialRunSpan := startTrace(ctx, "session.initial_run")

	temporalclient.WithoutDeadlockDetection(
		wctx,
		func() {
			run, err = sdkruntimes.Run(
				initialRunCtx,
				sdkruntimes.RunParams{
					Runtimes:             w.ws.svcs.Runtimes,
					BuildFile:            w.data.BuildFile,
					Globals:              w.globals,
					RunID:                runID,
					FallthroughCallbacks: cbs,
					EntryPointPath:       entryPoint.Path(),
					SessionID:            session.ID(),
					IsDurable:            session.IsDurable(),
				},
			)
		},
	)

	initialRunSpan.End()

	if err != nil {
		return printer.Finalize(), sdktypes.InvalidValue, err
	}

	kittehs.Must0(w.executors.AddExecutor(fmt.Sprintf("run_%s", run.ID().Value()), run))

	// Run call only if the entrypoint includes a name.
	if epName := entryPoint.Name(); epName != "" {
		callValue, ok := run.Values()[epName]
		if !ok {
			// The user specified an entry point that does not exist.
			// WrapError so it will be a program error and not considered as an internal error.
			return printer.Finalize(), sdktypes.InvalidValue, sdktypes.WrapError(fmt.Errorf("entry point %q not found after evaluation", epName)).ToError()
		}

		if !callValue.IsFunction() {
			// The user specified an entry point that is not a function.
			// WrapError so it will be a program error and not considered as an internal error.
			return printer.Finalize(), sdktypes.InvalidValue, sdktypes.WrapError(fmt.Errorf("entry point %q is not a function", epName)).ToError()
		}

		if callValue.GetFunction().ExecutorID().ToRunID() != runID {
			return printer.Finalize(), sdktypes.InvalidValue, errors.New("entry point does not belong to main run")
		}

		if err := w.updateState(wctx, sdktypes.NewSessionStateRunning(runID, callValue)); err != nil {
			return printer.Finalize(), sdktypes.InvalidValue, err
		}

		inputs := map[string]sdktypes.Value{
			"data":       sdktypes.NewDictValueFromStringMap(session.Inputs()),
			"session_id": sdktypes.NewStringValue(session.ID().String()),
		}

		callCtx, callSpan := startTrace(ctx, "session.call")
		callSpan.SetAttributes(attribute.String("function_name", callValue.GetFunction().Name().String()))

		if session.IsDurable() {
			// A durable session should be told by the runner when to start activities.
			// Here we just call directly the runner from the workflow itself.
			retVal, err = run.Call(callCtx, callValue, nil, inputs)
		} else {
			// A non-durable session is running entirely inside a single activity.
			retVal, err = w.call(wctx, runID, callValue, nil, inputs)
		}

		if err != nil {
			return printer.Finalize(), sdktypes.InvalidValue, err
		}

		callSpan.End()
	}

	prints := printer.Finalize()

	return prints, retVal, w.updateState(
		wctx,
		sdktypes.NewSessionStateCompleted(
			kittehs.Transform(prints, func(p sdkservices.SessionPrint) string { return p.Value.GetString().Value() }),
			run.Values(),
			retVal,
		))
}
