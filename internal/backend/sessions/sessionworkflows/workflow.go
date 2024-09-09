package sessionworkflows

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	httpmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/http"
	osmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/store"
	timemodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/time"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
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

	limitedTimeout = 5 * time.Second
)

func withLimitedTimeout(ctx context.Context) (context.Context, func()) {
	return context.WithTimeout(ctx, limitedTimeout)
}

var envVarsExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID(envVarsModuleName))

type sessionWorkflow struct {
	z  *zap.Logger
	ws *workflows

	data *sessiondata.Data

	debug bool

	// All the members belows must be built deterministically by the workflow.
	// They are not persisted in the database.

	executors sdkexecutor.Executors
	globals   map[string]sdktypes.Value

	callSeq uint32

	lastReadEventSeqForSignal map[uuid.UUID]uint64 // map signals to last read event seq num.

	state sdktypes.SessionState
}

type connInfo struct {
	Config          map[string]string `json:"config"`
	IntegrationName string            `json:"integration_name"`
}

func runWorkflow(
	ctx workflow.Context,
	z *zap.Logger,
	ws *workflows,
	data *sessiondata.Data,
	debug bool,
) (prints []string, err error) {
	w := &sessionWorkflow{
		z:                         z,
		data:                      data,
		ws:                        ws,
		debug:                     debug,
		lastReadEventSeqForSignal: make(map[uuid.UUID]uint64),
	}

	var cinfos map[string]connInfo

	if cinfos, err = w.initConnections(ctx); err != nil {
		return
	}

	if err = w.initEnvModule(cinfos); err != nil {
		return
	}

	if w.globals, err = w.initGlobalModules(); err != nil {
		return
	}

	prints, err = w.run(ctx)

	w.cleanupSignals(ctx)
	return
}

func (w *sessionWorkflow) cleanupSignals(ctx workflow.Context) {
	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)
	for signalID := range w.lastReadEventSeqForSignal {
		if err := w.ws.svcs.DB.RemoveSignal(goCtx, signalID); err != nil {
			// No need to to any handling in case of an error, it won't be used again
			// at most we would have db garabge we can clear up later with background jobs
			w.z.Warn(fmt.Sprintf("failed removing signal %s, err: %s", signalID, err), zap.Any("signalID", signalID), zap.Error(err))
		}
	}
}

func (w *sessionWorkflow) updateState(ctx workflow.Context, state sdktypes.SessionState) error {
	w.z.Debug("update state", zap.Any("state", state))

	w.state = state

	return workflow.ExecuteLocalActivity(ctx, w.ws.svcs.DB.UpdateSessionState, w.data.SessionID, state).Get(ctx, nil)
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

func (w *sessionWorkflow) call(ctx workflow.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if f := v.GetFunction(); f.HasFlag(sdktypes.ConstFunctionFlag) {
		return f.ConstValue()
	}

	w.callSeq++

	z := w.z.With(zap.Any("run_id", runID), zap.Any("v", v), zap.Uint32("seq", w.callSeq))

	z.Debug("call requested")

	result, err := w.ws.calls.Call(ctx, &sessioncalls.CallParams{
		SessionID: w.data.SessionID,
		CallSpec:  sdktypes.NewSessionCallSpec(v, args, kwargs, w.callSeq),
		Debug:     w.debug,
		Executors: &w.executors, // HACK
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

func (w *sessionWorkflow) initConnections(ctx workflow.Context) (map[string]connInfo, error) {
	cinfos := make(map[string]connInfo, len(w.data.Connections))

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	for _, conn := range w.data.Connections {
		name, cid, iid := conn.Name().String(), conn.ID(), conn.IntegrationID()

		intg, err := w.ws.svcs.Integrations.GetByID(goCtx, iid)
		if err != nil {
			return nil, fmt.Errorf("get integration %q: %w", iid, err)
		}

		if intg == nil {
			return nil, fmt.Errorf("integration %q not found", iid)
		}

		// In modules, we register the connection prefixed with its integration name.
		// This allows us to query all connections for a given integration in the load callback.
		scope := integrationModulePrefix(intg.Get().UniqueName().String()) + name
		if w.executors.GetValues(scope) != nil {
			return nil, fmt.Errorf("conflicting connection %q", name)
		}

		if xid := sdktypes.NewExecutorID(iid); w.executors.GetCaller(xid) == nil {
			kittehs.Must0(w.executors.AddCaller(xid, intg))
		}

		cinfo := connInfo{
			IntegrationName: intg.Get().UniqueName().String(),
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
	if err := sdktypes.VerifyEventFilter(filter); err != nil {
		w.z.Debug("invalid filter in workflow code", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("invalid filter: %w", err)
	}

	workflowID := workflow.GetInfo(wctx).WorkflowExecution.ID

	var minSequence uint64
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.GetLatestEventSequence).Get(wctx, &minSequence); err != nil {
		return uuid.UUID{}, fmt.Errorf("get current sequence: %w", err)
	}

	var signalID uuid.UUID
	if err := workflow.SideEffect(wctx, func(wctx workflow.Context) any {
		return uuid.New()
	}).Get(&signalID); err != nil {
		return uuid.UUID{}, fmt.Errorf("generate signal id: %w", err)
	}

	// must be set before signal is saved, otherwise the signal might reach the workflow before
	// the map is read.
	w.lastReadEventSeqForSignal[signalID] = minSequence

	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.SaveSignal, signalID, workflowID, did, filter).Get(wctx, nil); err != nil {
		return uuid.UUID{}, fmt.Errorf("save signal: %w", err)
	}

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

	// this will wait for first signal or timeout.
	selector.Select(wctx)

	return signalID, nil
}

func (w *sessionWorkflow) getNextEvent(ctx context.Context, signalID uuid.UUID) (map[string]sdktypes.Value, error) {
	wctx := sessioncontext.GetWorkflowContext(ctx)

	var signal scheme.Signal
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.GetSignal, signalID).Get(wctx, &signal); err != nil {
		w.z.Panic("get signal", zap.Error(err))
	}

	var did sdktypes.EventDestinationID
	if signal.ConnectionID != nil {
		did = sdktypes.NewEventDestinationID(sdktypes.NewIDFromUUID[sdktypes.ConnectionID](signal.ConnectionID))
	} else if signal.TriggerID != nil {
		did = sdktypes.NewEventDestinationID(sdktypes.NewIDFromUUID[sdktypes.TriggerID](signal.TriggerID))
	} else {
		return nil, fmt.Errorf("invalid signal %v: no connection or trigger ids, got %v", signalID, signal.DestinationID)
	}

	minSequenceNumber, ok := w.lastReadEventSeqForSignal[signalID]
	if !ok {
		return nil, fmt.Errorf("no such subscription %q", signalID)
	}

	filter := sdkservices.ListEventsFilter{
		DestinationID:     did,
		Limit:             1,
		MinSequenceNumber: minSequenceNumber + 1,
		Order:             sdkservices.ListOrderAscending,
	}

	var eventID sdktypes.EventID
	if err := workflow.ExecuteLocalActivity(wctx, func(ctx context.Context) (sdktypes.EventID, error) {
		for {
			evs, err := w.ws.svcs.DB.ListEvents(ctx, filter)
			if err != nil {
				return sdktypes.InvalidEventID, err
			}

			if len(evs) == 0 {
				return sdktypes.InvalidEventID, nil
			}

			ev, err := w.ws.svcs.DB.GetEventByID(ctx, evs[0].ID())
			if err != nil {
				return sdktypes.InvalidEventID, err
			}

			filter.MinSequenceNumber = ev.Seq() + 1

			match, err := ev.Matches(signal.Filter)
			if err != nil {
				// TODO(ENG-566): inform user.
				w.z.Info("invalid signal filter", zap.Error(err), zap.String("filter", signal.Filter))
				continue
			}

			if match {
				return ev.ID(), nil
			}
		}
	}).Get(wctx, &eventID); err != nil {
		w.z.Panic("get signal", zap.Error(err))
	}

	if !eventID.IsValid() {
		return nil, nil
	}

	event, err := w.ws.svcs.DB.GetEventByID(ctx, eventID)
	if err != nil {
		w.z.Panic("get event", zap.Error(err))
	}

	w.lastReadEventSeqForSignal[signalID] = event.Seq()

	return event.Data(), nil
}

func (w *sessionWorkflow) removeEventSubscription(ctx context.Context, signalID uuid.UUID) {
	wctx := sessioncontext.GetWorkflowContext(ctx)

	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.RemoveSignal, signalID).Get(wctx, nil); err != nil {
		w.z.Panic("remove signal", zap.Error(err))
	}

	delete(w.lastReadEventSeqForSignal, signalID)
}

func (w *sessionWorkflow) run(wctx workflow.Context) (prints []string, err error) {
	type contextKey string
	workflowContextKey := contextKey("autokitteh_workflow_context")

	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)

	// This will allow us to identify if the call context is from a workflow (script code run), or
	// some other thing that calls the Call callback from within an activity. The latter is not supported.
	ctx = context.WithValue(ctx, workflowContextKey, wctx)

	// TODO: replace with activity.IsActivity
	isFromActivity := func(ctx context.Context) bool { return ctx.Value(workflowContextKey) == nil }

	newRunID := func() (runID sdktypes.RunID) {
		if err := workflow.SideEffect(wctx, func(workflow.Context) any {
			return sdktypes.NewRunID()
		}).Get(&runID); err != nil {
			w.z.Panic("new run id side effect", zap.Error(err))
		}
		return
	}

	var run sdkservices.Run

	cbs := sdkservices.RunCallbacks{
		NewRunID: newRunID,
		Load:     w.load,
		Call: func(callCtx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			if run == nil {
				return sdktypes.InvalidValue, fmt.Errorf("cannot call before the run is initialized")
			}

			if xid := v.GetFunction().ExecutorID(); xid.ToRunID() == runID && w.executors.GetCaller(xid) == nil {
				// This happens only during initial evaluation (the first run because invoking the entrypoint function),
				// and the runtime tries to call itself in order to start an activity with its own functions.
				return sdktypes.InvalidValue, fmt.Errorf("cannot call self during initial evaluation")
			}

			if isFromActivity(callCtx) {
				return sdktypes.InvalidValue, fmt.Errorf("nested activities are not supported")
			}

			return w.call(wctx, runID, v, args, kwargs)
		},
		Print: func(printCtx context.Context, runID sdktypes.RunID, text string) {
			w.z.Debug("print", zap.String("run_id", runID.String()), zap.String("text", text))

			prints = append(prints, text)

			if isFromActivity(printCtx) || run == nil {
				// TODO: We do this since we're already in an activity. We need to either
				//       manually retry this (maybe even using the heartbeat to know if it's
				//       already been processed), or we need to aggregate the prints
				//       and just perform them in the workflow after the activity is done.
				err = w.ws.svcs.DB.AddSessionPrint(printCtx, w.data.SessionID, text)
			} else {
				err = workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.AddSessionPrint, w.data.SessionID, text).Get(wctx, nil)
			}

			if err != nil {
				w.z.Error("failed to add print session record", zap.String("run_id", runID.String()), zap.String("text", text))
			}
		},
	}

	runID := newRunID()

	if err := w.updateState(wctx, sdktypes.NewSessionStateRunning(runID, sdktypes.InvalidValue)); err != nil {
		return prints, err
	}

	entryPoint := w.data.Session.EntryPoint()

	type runOrErr struct {
		run sdkservices.Run
		err error
	}

	runDone := make(chan runOrErr)

	go func() {
		// This can take a while to complete since it might provision resources et al,
		// but since it might call back into the workflow via the callbacks, it must
		// still run a in a workflow context rather than an activity.
		// TODO: Consider just running this from an activity.
		run, err := sdkruntimes.Run(
			ctx,
			sdkruntimes.RunParams{
				Runtimes:             w.ws.svcs.Runtimes,
				BuildFile:            w.data.BuildFile,
				Globals:              w.globals,
				RunID:                runID,
				FallthroughCallbacks: cbs,
				EntryPointPath:       entryPoint.Path(),
			},
		)

		runDone <- runOrErr{run, err}
	}()

	for run == nil {
		select {
		case <-time.After(workflowDeadlockTimeout / 2):
			if err := workflow.Sleep(wctx, time.Millisecond); err != nil {
				return prints, err
			}
		case r := <-runDone:
			if r.err != nil {
				// TODO(ENG-130): discriminate between infra and real errors.
				return prints, r.err
			}

			run = r.run
		case <-ctx.Done():
			return prints, ctx.Err()
		}
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

		argNames := callValue.GetFunction().ArgNames() // sdktypes.GetFunctionValueArgsNames(callValue)
		kwargs := kittehs.FilterMapKeys(
			w.data.Session.Inputs(),
			kittehs.ContainedIn(argNames...),
		)

		if retVal, err = run.Call(ctx, callValue, nil, kwargs); err != nil {
			return prints, err
		}
	}

	state := sdktypes.NewSessionStateCompleted(prints, run.Values(), retVal)

	return prints, w.updateState(wctx, state)
}
