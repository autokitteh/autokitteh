package sessionworkflows

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	httpmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/http"
	osmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/store"
	testtoolsmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/testtools"
	timemodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/time"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
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

var envVarsExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID(envVarsModuleName))

type sessionWorkflow struct {
	l  *zap.Logger
	ws *workflows

	data *sessiondata.Data

	// All the members belows must be built deterministically by the workflow.
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
	data *sessiondata.Data,
) (prints []string, err error) {
	w := &sessionWorkflow{
		l:                         l,
		data:                      data,
		ws:                        ws,
		lastReadEventSeqForSignal: make(map[uuid.UUID]uint64),
	}

	var cinfos map[string]connInfo

	if cinfos, err = w.initConnections(wctx); err != nil {
		return
	}

	if err = w.initEnvModule(cinfos); err != nil {
		return
	}

	if w.globals, err = w.initGlobalModules(); err != nil {
		return
	}

	prints, err = w.run(wctx, l)

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
			// No need to to any handling in case of an error, it won't be used again
			// at most we would have db garabge we can clear up later with background jobs
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

func (w *sessionWorkflow) initGlobalModules() (map[string]sdktypes.Value, error) {
	execs := map[string]sdkexecutor.Executor{
		"ak":    w.newModule(),
		"time":  timemodule.New(),
		"http":  httpmodule.New(),
		"store": store.New(w.data.Env.ID(), w.data.ProjectID, w.ws.svcs.RedisClient),
	}

	vs := make(map[string]sdktypes.Value, len(execs))

	if w.ws.cfg.OSModule {
		execs["os"] = osmodule.New()
	} else {
		vs["os"] = sdktypes.Nothing
	}

	if w.ws.cfg.Test {
		execs["testtools"] = testtoolsmodule.New()
	} else {
		vs["testtools"] = sdktypes.Nothing
	}

	for name, exec := range execs {
		sym, err := sdktypes.StrictParseSymbol(name)
		if err != nil {
			return nil, err
		}
		vs[name] = kittehs.Must1(sdktypes.NewStructValue(sdktypes.NewSymbolValue(sym), exec.Values()))
		if err := w.executors.AddExecutor(name, exec); err != nil {
			return nil, err
		}
	}

	return vs, nil
}

func (w *sessionWorkflow) createEventSubscription(wctx workflow.Context, filter string, did sdktypes.EventDestinationID) (uuid.UUID, error) {
	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, w.ws.cfg.Activity)

	sl := w.l.Sugar().With("destination_id", did)

	if err := sdktypes.VerifyEventFilter(filter); err != nil {
		sl.With("err", err).Infof("invalid filter in workflow code: %v", err)
		return uuid.Nil, sdkerrors.NewInvalidArgumentError("invalid filter: %w", err)
	}

	// generate a unique signal id.
	var signalID uuid.UUID
	if err := workflow.SideEffect(wctx, func(wctx workflow.Context) any {
		return uuid.New()
	}).Get(&signalID); err != nil {
		return uuid.Nil, fmt.Errorf("generate signal id: %w", err)
	}

	var minSequence uint64
	if err := workflow.ExecuteActivity(wctx, getLastEventSequenceActivityName).Get(wctx, &minSequence); err != nil {
		return uuid.Nil, fmt.Errorf("get current sequence: %w", err)
	}

	// must be set before signal is saved, otherwise the signal might reach the workflow before
	// the map is updated.
	w.lastReadEventSeqForSignal[signalID] = minSequence

	workflowID := workflow.GetInfo(wctx).WorkflowExecution.ID

	signal := types.Signal{
		ID:            signalID,
		WorkflowID:    workflowID,
		DestinationID: did,
		Filter:        filter,
	}

	if err := workflow.ExecuteActivity(wctx, saveSignalActivityName, &signal).Get(wctx, nil); err != nil {
		return uuid.Nil, fmt.Errorf("save signal: %w", err)
	}

	sl.Info("created event subscription %v", signalID)

	return signalID, nil
}

// Returns "", nil on timeout.
func (w *sessionWorkflow) waitOnFirstSignal(wctx workflow.Context, signals []uuid.UUID, f workflow.Future) (uuid.UUID, error) {
	selector := workflow.NewSelector(wctx)

	if f != nil {
		selector.AddFuture(f, func(workflow.Future) {})
	}

	var signalID uuid.UUID

	for _, signal := range signals {
		selector.AddReceive(workflow.GetSignalChannel(wctx, signal.String()), func(c workflow.ReceiveChannel, _ bool) {
			// clear all pending signals.
			for c.ReceiveAsync(nil) {
				// nop
			}

			signalID = signal
		})
	}

	var cancelled bool

	// Select doesn't respond to cancellations unless we add a receive on the context done channel.
	selector.AddReceive(wctx.Done(), func(c workflow.ReceiveChannel, _ bool) { cancelled = true })

	// this will wait for first signal or timeout.
	selector.Select(wctx)

	if cancelled {
		return uuid.Nil, wctx.Err()
	}

	return signalID, nil
}

func (w *sessionWorkflow) getNextEvent(ctx context.Context, sigid uuid.UUID) (map[string]sdktypes.Value, error) {
	sl := w.l.Sugar().With("signal_id", sigid)

	wctx := sessioncontext.GetWorkflowContext(ctx)
	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, w.ws.cfg.Activity)

	minSequenceNumber, ok := w.lastReadEventSeqForSignal[sigid]
	if !ok {
		return nil, fmt.Errorf("no such subscription %q", sigid)
	}

	var event sdktypes.Event

	fut := workflow.ExecuteActivity(
		wctx,
		getSignalEventActivityName,
		sigid,
		minSequenceNumber,
	)

	if err := fut.Get(wctx, &event); err != nil {
		// was the context cancelled?
		if wctx.Err() != nil {
			return nil, err
		}

		return nil, fmt.Errorf("get signal event %v: %w", sigid, err)
	}

	if !event.IsValid() {
		return nil, nil
	}

	w.lastReadEventSeqForSignal[sigid] = event.Seq()

	sl.With("event_id", event.ID()).Infof("got event %v", event.ID())

	return event.Data(), nil
}

func (w *sessionWorkflow) removeEventSubscription(ctx context.Context, signalID uuid.UUID) {
	sl := w.l.Sugar().With("signal_id", signalID)

	wctx := sessioncontext.GetWorkflowContext(ctx)
	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, w.ws.cfg.Activity)

	if err := workflow.ExecuteActivity(wctx, removeSignalActivityName, signalID).Get(wctx, nil); err != nil {
		// it is not a critical error, we can just log it. no need to panic.
		sl.With(err).Errorf("remove signal: %v", signalID, err)
	}

	delete(w.lastReadEventSeqForSignal, signalID)
}

func (w *sessionWorkflow) run(wctx workflow.Context, l *zap.Logger) (prints []string, err error) {
	sl := l.Sugar()

	sid := w.data.Session.ID()

	newRunID := func() (runID sdktypes.RunID) {
		if err := workflow.SideEffect(wctx, func(workflow.Context) any {
			return sdktypes.NewRunID()
		}).Get(&runID); err != nil {
			sl.With("err", err).Panicf("new run id side effect: %v", err)
		}
		return
	}

	var run sdkservices.Run

	cbs := sdkservices.RunCallbacks{
		NewRunID: newRunID,
		Load:     w.load,
		Call: func(callCtx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			if f := v.GetFunction(); f.HasFlag(sdktypes.ConstFunctionFlag) {
				sl.Debug("const function call")
				return f.ConstValue()
			}

			isActivity := activity.IsActivity(callCtx)

			sl := sl.With("run_id", runID, "is_activity", isActivity, "v", v)

			if run == nil {
				sl.Debug("run not initialized")
				return sdktypes.InvalidValue, fmt.Errorf("cannot call before the run is initialized")
			}

			if xid := v.GetFunction().ExecutorID(); xid.ToRunID() == runID && w.executors.GetCaller(xid) == nil {
				sl.Debug("self call during initialization")

				// This happens only during initial evaluation (the first run because invoking the entrypoint function),
				// and the runtime tries to call itself in order to start an activity with its own functions.
				return sdktypes.InvalidValue, fmt.Errorf("cannot call self during initial evaluation")
			}

			if isActivity {
				sl.Debug("nested activity call")
				return sdktypes.InvalidValue, fmt.Errorf("nested activities are not supported")
			}

			return w.call(wctx, runID, v, args, kwargs)
		},
		Print: func(printCtx context.Context, runID sdktypes.RunID, text string) {
			isActivity := activity.IsActivity(printCtx)

			// Trim single trailing space, but no other spaces.
			if text[len(text)-1] == '\n' {
				text = text[:len(text)-1]
			}

			sl := sl.With("run_id", runID, "is_activity", isActivity)

			sl.Debugw("print", zap.String("text", text))

			prints = append(prints, text)

			if run == nil || isActivity {
				// run == nil: Since initial run is running via temporalclient.LongRunning, we
				// cannot use the workflow context to report prints.
				//
				// TODO(ENG-1554): we need to retry here.
				err = w.ws.svcs.DB.AddSessionPrint(printCtx, sid, text)
			} else {
				// TODO(ENG-1467): Currently the python runtime is calling Print from a
				//                 separate goroutine. This is a workaround to detect
				//                 that behaviour and in that case just write the
				//                 print directly into the DB since we cannot launch
				//                 an activity from a separate goroutine.
				if temporalclient.IsWorkflowContextAsGoContext(printCtx) {
					err = workflow.ExecuteActivity(wctx, addSessionPrintActivityName, sid, text).Get(wctx, nil)
				} else if !workflow.IsReplaying(wctx) {
					err = w.ws.svcs.DB.AddSessionPrint(printCtx, sid, text)
				}
			}

			// We do not consider print failure as a critical error, since we don't want to hold back the
			// workflow for potential user debugging prints. Just log the error and move on. Nevertheless,
			// this is a problem to be aware of, because errors cause the loss of valuable debugging data
			// (because the workflow context is canceled).
			if err != nil {
				sl.With("text", text).Warnf("failed to add print session record: %v", err)
			}
		},
	}

	runID := newRunID()

	if err := w.updateState(wctx, sdktypes.NewSessionStateRunning(runID, sdktypes.InvalidValue)); err != nil {
		return nil, err
	}

	entryPoint := w.data.Session.EntryPoint()

	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	temporalclient.WithoutDeadlockDetection(
		wctx,
		func() {
			run, err = sdkruntimes.Run(
				ctx,
				sdkruntimes.RunParams{
					Runtimes:             w.ws.svcs.Runtimes,
					BuildFile:            w.data.BuildFile,
					Globals:              w.globals,
					RunID:                runID,
					FallthroughCallbacks: cbs,
					EntryPointPath:       entryPoint.Path(),
					SessionID:            w.data.Session.ID(),
				},
			)
		},
	)

	if err != nil {
		return nil, err
	}

	kittehs.Must0(w.executors.AddExecutor(fmt.Sprintf("run_%s", run.ID().Value()), run))

	var retVal sdktypes.Value

	// Run call only if the entrypoint includes a name.
	if epName := entryPoint.Name(); epName != "" {
		callValue, ok := run.Values()[epName]
		if !ok {
			return prints, fmt.Errorf("entry point not found after evaluation")
		}

		if !callValue.IsFunction() {
			return prints, fmt.Errorf("entry point is not a function")
		}

		if callValue.GetFunction().ExecutorID().ToRunID() != runID {
			return prints, fmt.Errorf("entry point does not belong to main run")
		}

		if err := w.updateState(wctx, sdktypes.NewSessionStateRunning(runID, callValue)); err != nil {
			return prints, err
		}

		argNames := callValue.GetFunction().ArgNames()
		kwargs := kittehs.FilterMapKeys(w.data.Session.Inputs(), kittehs.ContainedIn(argNames...))

		if retVal, err = run.Call(ctx, callValue, nil, kwargs); err != nil {
			return prints, err
		}
	}

	return prints, w.updateState(wctx, sdktypes.NewSessionStateCompleted(prints, run.Values(), retVal))
}
