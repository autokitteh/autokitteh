package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	httpmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/http"
	osmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/store"
	testtoolsmodule "go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/testtools"
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
		"store": store.New(w.data.Session.ProjectID(), w.ws.svcs.RedisClient),
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

func (w *sessionWorkflow) run(wctx workflow.Context, l *zap.Logger) (prints []string, err error) {
	var run sdkservices.Run

	cbs := runCallbacks{
		wctx:  wctx,
		w:     w,
		l:     l,
		run:   func() sdkservices.Run { return run },
		print: func(text string) { prints = append(prints, text) },
	}

	runID := cbs.NewRunID()

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
			// The user specified an entry point that does not exist.
			// WrapError so it will be a program error and not considered as an internal error.
			return prints, sdktypes.WrapError(fmt.Errorf("entry point %q not found after evaluation", epName)).ToError()
		}

		if !callValue.IsFunction() {
			// The user specified an entry point that is not a function.
			// WrapError so it will be a program error and not considered as an internal error.
			return prints, sdktypes.WrapError(fmt.Errorf("entry point %q is not a function", epName)).ToError()
		}

		if callValue.GetFunction().ExecutorID().ToRunID() != runID {
			return prints, errors.New("entry point does not belong to main run")
		}

		if err := w.updateState(wctx, sdktypes.NewSessionStateRunning(runID, callValue)); err != nil {
			return prints, err
		}

		if retVal, err = run.Call(ctx, callValue, nil, w.data.Session.Inputs()); err != nil {
			return prints, err
		}
	}

	return prints, w.updateState(wctx, sdktypes.NewSessionStateCompleted(prints, run.Values(), retVal))
}
