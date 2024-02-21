package sessionworkflows

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/akmodules/ak"
	"go.autokitteh.dev/autokitteh/backend/internal/akmodules/store"
	timemodule "go.autokitteh.dev/autokitteh/backend/internal/akmodules/time"
	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/backend/internal/fixtures"
	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessiondata"
	"go.autokitteh.dev/autokitteh/backend/internal/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkrun"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	envVarsModuleName     = "env"
	integrationPathPrefix = "@"
)

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

	poller sdktypes.Value
	fakers map[string]sdktypes.Value

	callSeq uint32

	signals map[string]uint64 // map signals to next sequence number
}

func runWorkflow(
	ctx workflow.Context,
	z *zap.Logger,
	ws *workflows,
	data *sessiondata.Data,
	debug bool,
) error {
	w := &sessionWorkflow{
		z:       z,
		data:    data,
		poller:  sdktypes.NewNothingValue(),
		ws:      ws,
		fakers:  make(map[string]sdktypes.Value),
		debug:   debug,
		signals: make(map[string]uint64),
	}

	w.initEnvModule()

	var err error
	if w.globals, err = w.initGlobalModules(ctx); err != nil {
		w.updateState(ctx, sdktypes.NewErrorSessionState(err, nil))
		return err // definitely an infra error.
	}

	if err := w.initConnections(ctx); err != nil {
		w.updateState(ctx, sdktypes.NewErrorSessionState(err, nil))
		return nil // not an infra error.
	}

	if prints, err := w.run(ctx); err != nil {
		w.updateState(ctx, sdktypes.NewErrorSessionState(err, prints))
		return nil // not an infra error.
	}

	return nil
}

func (w *sessionWorkflow) addPrint(ctx workflow.Context, print string) {
	w.z.Debug("add print", zap.String("print", print))

	if err := workflow.ExecuteLocalActivity(ctx, w.ws.svcs.DB.AddSessionPrint, w.data.SessionID, print).Get(ctx, nil); err != nil {
		w.z.Panic("add print", zap.Error(err))
	}
}

func (w *sessionWorkflow) updateState(ctx workflow.Context, state sdktypes.Object) {
	wrapped := sdktypes.SessionStateWithTimestamp(sdktypes.WrapSessionState(state), time.Now())

	w.z.Debug("update state", zap.Any("state", wrapped))

	if err := workflow.ExecuteLocalActivity(ctx, w.ws.svcs.DB.UpdateSessionState, w.data.SessionID, wrapped).Get(ctx, nil); err != nil {
		w.z.Panic("update session", zap.Error(err))
	}
}

func (w *sessionWorkflow) loadIntegrationConnections(ctx context.Context, path string) (map[string]sdktypes.Value, error) {
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
				return "", nil, fmt.Errorf("invalid symbol %q: %w", n, err)
			}

			return n, sdktypes.NewStructValue(sdktypes.NewSymbolValue(sym), vs), nil
		},
	)
}

func (w *sessionWorkflow) load(ctx context.Context, _ sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
	if strings.HasPrefix(path, integrationPathPrefix) {
		return w.loadIntegrationConnections(ctx, path[1:])
	}

	vs := w.executors.GetValues(path)
	if vs == nil {
		return nil, sdkerrors.ErrNotFound
	}

	return vs, nil
}

func (w *sessionWorkflow) call(ctx workflow.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	w.callSeq++

	z := w.z.With(zap.Any("run_id", runID), zap.Any("v", v), zap.Uint32("seq", w.callSeq))

	z.Debug("call requested")

	// TODO: make sure that fake is pure starlark.
	callvID := sdktypes.GetFunctionValueUniqueID(v)
	fake, useFake := w.fakers[callvID]
	if useFake {
		z.With(zap.Any("fake", fake))
		v = fake
		z.Debug("using fake")
	}

	result, err := w.ws.calls.Call(ctx, &sessioncalls.CallParams{
		SessionID:     w.data.SessionID,
		CallSpec:      sdktypes.NewSessionCallSpec(v, args, kwargs, w.callSeq),
		Debug:         w.debug,
		ForceInternal: useFake,
		Poller:        w.getPoller(),
		Executors:     &w.executors, // HACK
	})
	if err != nil {
		z.Panic("call", zap.Error(err))
	}

	return sdktypes.SessionCallResultAsPair(result)
}

func (w *sessionWorkflow) initEnvModule() {
	mod := sdkexecutor.NewExecutor(
		nil, // no calls will be ever made to env.
		envVarsExecutorID,
		kittehs.ListToMap(w.data.EnvVars, func(v sdktypes.EnvVar) (string, sdktypes.Value) {
			return sdktypes.GetEnvVarName(v).String(), sdktypes.NewStringValue(sdktypes.GetEnvVarValue(v))
		}),
	)

	kittehs.Must0(w.executors.AddExecutor(envVarsModuleName, mod))
}

func integrationModulePrefix(name string) string { return fmt.Sprintf("__%v__", name) }

func (w *sessionWorkflow) initConnections(ctx workflow.Context) error {
	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	for _, conn := range w.data.Connections {
		name := sdktypes.GetConnectionName(conn).String()
		iid := sdktypes.GetConnectionIntegrationID(conn)

		if w.executors.GetValues(name) != nil {
			return fmt.Errorf("conflicting connection %q", name)
		}

		// In modules, we register the connection prefixed with its integration name.
		// This allows us to query all connections for a given integration in the load callback.

		intg, err := w.ws.svcs.Integrations.Get(goCtx, iid)
		if err != nil {
			return fmt.Errorf("get integration %q: %w", iid, err)
		}

		if intg == nil {
			return fmt.Errorf("integration %q not found", iid)
		}

		if xid := sdktypes.NewExecutorID(iid); w.executors.GetCaller(xid) == nil {
			kittehs.Must0(w.executors.AddCaller(xid, intg))
		}

		// mod's executor id is the integration id.
		vs, err := intg.Configure(goCtx, sdktypes.GetConnectionIntegrationToken(conn))
		if err != nil {
			return fmt.Errorf("connect to integration %q: %w", iid, err)
		}

		scope := integrationModulePrefix(sdktypes.GetIntegrationUniqueName(intg.Get()).String()) + name

		if err := w.executors.AddValues(scope, vs); err != nil {
			return err
		}
	}

	return nil
}

func (w *sessionWorkflow) initGlobalModules(ctx workflow.Context) (map[string]sdktypes.Value, error) {
	execs := map[string]sdkexecutor.Executor{
		"ak":    ak.New(w.syscall),
		"time":  timemodule.New(),
		"store": store.New(sdktypes.GetEnvID(w.data.Env), w.data.ProjectID, w.ws.svcs.RedisClient),
	}

	vs := make(map[string]sdktypes.Value, len(execs))

	for name, exec := range execs {
		sym, err := sdktypes.StrictParseSymbol(name)
		if err != nil {
			return nil, fmt.Errorf("invalid symbol %q: %w", name, err)
		}
		vs[name] = sdktypes.NewStructValue(sdktypes.NewSymbolValue(sym), exec.Values())
		if err := w.executors.AddExecutor(name, exec); err != nil {
			return nil, err
		}
	}

	return vs, nil
}

func (w *sessionWorkflow) createEventSubscription(ctx context.Context, connectionName, eventType string) (string, error) {
	wctx := sessioncontext.GetWorkflowContext(ctx)
	workflowID := workflow.GetInfo(wctx).WorkflowExecution.ID
	signalID := fmt.Sprintf("wid_%s_cn_%s_et_%s", workflowID, connectionName, eventType)

	_, connection := kittehs.FindFirst(w.data.Connections, func(c sdktypes.Connection) bool {
		return sdktypes.GetConnectionName(c).String() == connectionName
	})

	if connection == nil {
		return "", fmt.Errorf("connection %q not found", connectionName)
	}

	cid := sdktypes.GetConnectionID(connection)
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.SaveSignal, signalID, workflowID, cid, eventType).Get(wctx, nil); err != nil {
		w.z.Panic("save signal", zap.Error(err))
	}

	var c sdktypes.Connection
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.Connections.Get, cid).Get(wctx, &c); err != nil {
		w.z.Panic("get connection", zap.Error(err))
	}

	var minSequence uint64
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.GetLatestEventSequence).Get(wctx, &minSequence); err != nil {
		w.z.Panic("get current sequence", zap.Error(err))
	}

	w.signals[signalID] = minSequence

	return signalID, nil
}

func (w *sessionWorkflow) waitOnFirstSignal(ctx context.Context, signals []string) (string, error) {
	wctx := sessioncontext.GetWorkflowContext(ctx)
	selector := workflow.NewSelector(wctx)

	var signalID string
	for _, signal := range signals {
		func(s string) {
			selector.AddReceive(workflow.GetSignalChannel(wctx, s), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(wctx, nil) // we don't really care about the signal data
				signalID = s
			})
		}(signal)
	}

	// this will wait for first signal
	selector.Select(wctx)

	return signalID, nil
}

func (w *sessionWorkflow) getNextEvent(ctx context.Context, signalID string) (map[string]sdktypes.Value, error) {
	wctx := sessioncontext.GetWorkflowContext(ctx)

	var signal scheme.Signal
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.GetSignal, signalID).Get(wctx, &signal); err != nil {
		w.z.Panic("get signal", zap.Error(err))
	}

	iid := kittehs.Must1(sdktypes.ParseIntegrationID(signal.Connection.IntegrationID))

	minSequenceNumber, ok := w.signals[signalID]
	if !ok {
		w.z.Panic("signal not found")
	}

	filter := sdkservices.ListEventsFilter{
		IntegrationID:     iid,
		IntegrationToken:  signal.Connection.IntegrationToken,
		EventType:         signal.EventType,
		Limit:             1,
		MinSequenceNumber: minSequenceNumber,
	}

	var eventID sdktypes.EventID
	if err := workflow.ExecuteLocalActivity(wctx, func(ctx context.Context) (sdktypes.EventID, error) {
		evs, err := w.ws.svcs.DB.ListEvents(ctx, filter)
		if err != nil {
			return nil, err
		}

		if len(evs) == 0 {
			return nil, nil
		}

		return sdktypes.GetEventID(evs[0]), nil
	}).Get(wctx, &eventID); err != nil {
		w.z.Panic("get signal", zap.Error(err))
	}

	if eventID == nil {
		return nil, nil
	}

	event, err := w.ws.svcs.DB.GetEventByID(ctx, eventID)
	if err != nil {
		w.z.Panic("get event", zap.Error(err))
	}

	w.signals[signalID] = sdktypes.GetEventSequenceNumber(event) + 1

	return sdktypes.GetEventData(event), nil
}

func (w *sessionWorkflow) removeEventSubscription(ctx context.Context, signalID string) {
	wctx := sessioncontext.GetWorkflowContext(ctx)

	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.RemoveSignal, signalID).Get(wctx, nil); err != nil {
		w.z.Panic("remove signal", zap.Error(err))
	}

	delete(w.signals, signalID)
}

func (w *sessionWorkflow) run(ctx workflow.Context) (prints []string, err error) {
	// set before r.Call is called.
	var callValue sdktypes.Value

	newRunID := func() (runID sdktypes.RunID) {
		if err := workflow.SideEffect(ctx, func(workflow.Context) any {
			return sdktypes.NewRunID()
		}).Get(&runID); err != nil {
			w.z.Panic("new run id side effect", zap.Error(err))
		}
		return
	}

	cbs := sdkservices.RunCallbacks{
		NewRunID: newRunID,
		Load:     w.load,
		Call: func(_ context.Context, rid sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			return w.call(ctx, rid, v, args, kwargs)
		},
		Print: func(_ context.Context, runID sdktypes.RunID, text string) {
			w.z.Debug("print", zap.String("run_id", runID.String()), zap.String("text", text))

			prints = append(prints, text)

			w.addPrint(ctx, text)
		},
	}

	runID := newRunID()

	w.updateState(ctx, sdktypes.NewRunningSessionState(runID, nil))

	entryPoint := sdktypes.GetSessionEntryPoint(w.data.Session)

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	run, err := sdkrun.Run(
		goCtx,
		sdkrun.RunParams{
			Runtimes:             w.ws.svcs.Runtimes,
			BuildFile:            w.data.BuildFile,
			Globals:              w.globals,
			RunID:                runID,
			FallthroughCallbacks: cbs,
			EntryPointPath:       sdktypes.GetCodeLocationPath(entryPoint),
		},
	)
	if err != nil {
		// TODO(ENG-130): discriminate between infra and real errors.
		return prints, err
	}

	kittehs.Must0(w.executors.AddExecutor(fmt.Sprintf("run_%s", run.ID().Value()), run))

	epName := sdktypes.GetCodeLocationName(entryPoint)

	callValue, ok := run.Values()[epName]
	if !ok {
		return prints, fmt.Errorf("entry point not found after evaluation")
	}

	if !sdktypes.IsFunctionValue(callValue) {
		return prints, fmt.Errorf("entry point is not a function")
	}

	if sdktypes.GetFunctionValueExecutorID(callValue).String() != runID.String() {
		return prints, fmt.Errorf("entry point does not belong to main run")
	}

	w.updateState(ctx, sdktypes.NewRunningSessionState(runID, callValue))

	argNames := sdktypes.GetFunctionValueArgsNames(callValue)
	kwargs := kittehs.FilterMapKeys(
		sdktypes.GetSessionInputs(w.data.Session),
		kittehs.ContainedIn(argNames...),
	)

	ret, err := run.Call(goCtx, callValue, nil, kwargs)
	if err != nil {
		return prints, err
	}

	state, err := sdktypes.NewCompletedSessionState(prints, run.Values(), ret)
	if err != nil {
		return prints, err
	}

	w.updateState(ctx, state)

	return prints, nil
}
