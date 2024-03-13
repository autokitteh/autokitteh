package sessionworkflows

import (
	"context"
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/ak"
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/store"
	timemodule "go.autokitteh.dev/autokitteh/internal/backend/akmodules/time"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
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
		poller:  sdktypes.Nothing,
		ws:      ws,
		fakers:  make(map[string]sdktypes.Value),
		debug:   debug,
		signals: make(map[string]uint64),
	}

	w.initEnvModule()

	var err error
	if w.globals, err = w.initGlobalModules(); err != nil {
		w.updateState(ctx, sdktypes.NewSessionStateError(err, nil))
		return err // definitely an infra error.
	}

	if err := w.initConnections(ctx); err != nil {
		w.updateState(ctx, sdktypes.NewSessionStateError(err, nil))
		return nil // not an infra error.
	}

	if prints, err := w.run(ctx); err != nil {
		w.updateState(ctx, sdktypes.NewSessionStateError(err, prints))
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

func (w *sessionWorkflow) updateState(ctx workflow.Context, state sdktypes.SessionState) {
	w.z.Debug("update state", zap.Any("state", state))

	if err := workflow.ExecuteLocalActivity(ctx, w.ws.svcs.DB.UpdateSessionState, w.data.SessionID, state).Get(ctx, nil); err != nil {
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
				return "", sdktypes.InvalidValue, fmt.Errorf("invalid symbol %q: %w", n, err)
			}

			return n, kittehs.Must1(sdktypes.NewStructValue(sdktypes.NewSymbolValue(sym), vs)), nil
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
	callvID := v.GetFunction().UniqueID()
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

	return result.ToPair()
}

func (w *sessionWorkflow) initEnvModule() {
	mod := sdkexecutor.NewExecutor(
		nil, // no calls will be ever made to env.
		envVarsExecutorID,
		kittehs.ListToMap(w.data.EnvVars, func(v sdktypes.EnvVar) (string, sdktypes.Value) {
			return v.Symbol().String(), sdktypes.NewStringValue(v.Value())
		}),
	)

	kittehs.Must0(w.executors.AddExecutor(envVarsModuleName, mod))
}

func integrationModulePrefix(name string) string { return fmt.Sprintf("__%v__", name) }

func (w *sessionWorkflow) initConnections(ctx workflow.Context) error {
	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	for _, conn := range w.data.Connections {
		name := conn.Name().String()
		iid := conn.IntegrationID()

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
		vs, err := intg.Configure(goCtx, conn.IntegrationToken())
		if err != nil {
			return fmt.Errorf("connect to integration %q: %w", iid, err)
		}

		scope := integrationModulePrefix(intg.Get().UniqueName().String()) + name

		if err := w.executors.AddValues(scope, vs); err != nil {
			return err
		}
	}

	return nil
}

func (w *sessionWorkflow) initGlobalModules() (map[string]sdktypes.Value, error) {
	execs := map[string]sdkexecutor.Executor{
		"ak":    ak.New(w.syscall),
		"time":  timemodule.New(),
		"store": store.New(w.data.Env.ID(), w.data.ProjectID, w.ws.svcs.RedisClient),
	}

	vs := make(map[string]sdktypes.Value, len(execs))

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

func (w *sessionWorkflow) createEventSubscription(ctx context.Context, connectionName, eventType string) (string, error) {
	wctx := sessioncontext.GetWorkflowContext(ctx)
	workflowID := workflow.GetInfo(wctx).WorkflowExecution.ID
	signalID := fmt.Sprintf("wid_%s_cn_%s_et_%s", workflowID, connectionName, eventType)

	_, connection := kittehs.FindFirst(w.data.Connections, func(c sdktypes.Connection) bool {
		return c.Name().String() == connectionName
	})

	if !connection.IsValid() {
		return "", fmt.Errorf("connection %q not found", connectionName)
	}

	cid := connection.ID()
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
			return sdktypes.InvalidEventID, err
		}

		if len(evs) == 0 {
			return sdktypes.InvalidEventID, nil
		}

		return evs[0].ID(), nil
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

	w.signals[signalID] = event.Seq() + 1

	return event.Data(), nil
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

	w.updateState(ctx, sdktypes.NewSessionStateRunning(runID, sdktypes.InvalidValue))

	entryPoint := w.data.Session.EntryPoint()

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	run, err := sdkruntimes.Run(
		goCtx,
		sdkruntimes.RunParams{
			Runtimes:             w.ws.svcs.Runtimes,
			BuildFile:            w.data.BuildFile,
			Globals:              w.globals,
			RunID:                runID,
			FallthroughCallbacks: cbs,
			EntryPointPath:       entryPoint.Path(),
		},
	)
	if err != nil {
		// TODO(ENG-130): discriminate between infra and real errors.
		return prints, err
	}

	kittehs.Must0(w.executors.AddExecutor(fmt.Sprintf("run_%s", run.ID().Value()), run))

	epName := entryPoint.Name()

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

	w.updateState(ctx, sdktypes.NewSessionStateRunning(runID, callValue))

	argNames := callValue.GetFunction().ArgNames() // sdktypes.GetFunctionValueArgsNames(callValue)
	kwargs := kittehs.FilterMapKeys(
		w.data.Session.Inputs(),
		kittehs.ContainedIn(argNames...),
	)

	ret, err := run.Call(goCtx, callValue, nil, kwargs)
	if err != nil {
		return prints, err
	}

	state := sdktypes.NewSessionStateCompleted(prints, run.Values(), ret)

	w.updateState(ctx, state)

	return prints, nil
}
