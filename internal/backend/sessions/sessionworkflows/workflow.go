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

	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/ak"
	osmodule "go.autokitteh.dev/autokitteh/internal/backend/akmodules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/store"
	timemodule "go.autokitteh.dev/autokitteh/internal/backend/akmodules/time"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
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

	poller sdktypes.Value
	fakers map[string]sdktypes.Value

	callSeq uint32

	signals map[string]uint64 // map signals to next sequence number

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
		z:       z,
		data:    data,
		poller:  sdktypes.Nothing,
		ws:      ws,
		fakers:  make(map[string]sdktypes.Value),
		debug:   debug,
		signals: make(map[string]uint64),
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

	return
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

	f := v.GetFunction()

	z.Debug("call requested")

	// TODO: make sure that fake is pure starlark?
	callvID := f.UniqueID()
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
		"ak":    ak.New(w.syscall, w.data, w.ws.svcs),
		"time":  timemodule.New(),
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

func (w *sessionWorkflow) createEventSubscription(ctx context.Context, connectionName, filter string) (string, error) {
	if err := sdktypes.VerifyEventFilter(filter); err != nil {
		w.z.Debug("invalid filter in workflow code", zap.Error(err))
		return "", fmt.Errorf("invalid filter: %w", err)
	}
	wctx := sessioncontext.GetWorkflowContext(ctx)
	workflowID := workflow.GetInfo(wctx).WorkflowExecution.ID

	var minSequence uint64
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.GetLatestEventSequence).Get(wctx, &minSequence); err != nil {
		return "", fmt.Errorf("get current sequence: %w", err)
	}

	signalID := uuid.New().String()

	_, connection := kittehs.FindFirst(w.data.Connections, func(c sdktypes.Connection) bool {
		return c.Name().String() == connectionName
	})

	if !connection.IsValid() {
		return "", fmt.Errorf("connection %q not found", connectionName)
	}

	cid := connection.ID()
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.DB.SaveSignal, signalID, workflowID, cid, filter).Get(wctx, nil); err != nil {
		return "", fmt.Errorf("save signal: %w", err)
	}

	var c sdktypes.Connection
	if err := workflow.ExecuteLocalActivity(wctx, w.ws.svcs.Connections.Get, cid).Get(wctx, &c); err != nil {
		return "", fmt.Errorf("get connection: %w", err)
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

	iid := sdktypes.NewIDFromUUID[sdktypes.IntegrationID](signal.Connection.IntegrationID)
	cid := sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&signal.Connection.ConnectionID)

	minSequenceNumber, ok := w.signals[signalID]
	if !ok {
		w.z.Panic("signal not found")
	}

	filter := sdkservices.ListEventsFilter{
		IntegrationID:     iid,
		ConnectionID:      cid,
		Limit:             1,
		MinSequenceNumber: minSequenceNumber,
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

func (w *sessionWorkflow) run(wctx workflow.Context) (prints []string, err error) {
	type contextKey string
	workflowContextKey := contextKey("autokitteh_workflow_context")

	goCtx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	// This will allow us to identify if the call context is from a workflow (script code run), or
	// some other thing that calls the Call callback from within an activity. The latter is not supported.
	goCtx = context.WithValue(goCtx, workflowContextKey, wctx)
	goCtx = authcontext.SetComponent(goCtx, "sessionWF")
	isFromActivity := func(ctx context.Context) bool { return ctx.Value(workflowContextKey) == nil }

	newRunID := func() (runID sdktypes.RunID) {
		if err := workflow.SideEffect(wctx, func(workflow.Context) any {
			return sdktypes.NewRunID()
		}).Get(&runID); err != nil {
			w.z.Panic("new run id side effect", zap.Error(err))
		}
		return
	}

	cbs := sdkservices.RunCallbacks{
		NewRunID: newRunID,
		Load:     w.load,
		Call: func(callCtx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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

			if isFromActivity(printCtx) {
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

		if retVal, err = run.Call(goCtx, callValue, nil, kwargs); err != nil {
			return prints, err
		}
	}

	state := sdktypes.NewSessionStateCompleted(prints, run.Values(), retVal)

	return prints, w.updateState(wctx, state)
}
