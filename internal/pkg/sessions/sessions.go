package sessions

import (
	"context"
	"fmt"
	"time"

	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/autokitteh/autokitteh/internal/pkg/akmod"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/plugin"
	"github.com/autokitteh/autokitteh/internal/pkg/pluginsreg"
	"github.com/autokitteh/autokitteh/internal/pkg/programs"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
	"github.com/autokitteh/autokitteh/sdk/api/apievent"
	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"

	"github.com/autokitteh/L"
)

type Config struct{}

type Sessions struct {
	Config      Config
	Temporal    temporalclient.Client
	EventsStore eventsstore.Store
	Programs    *programs.Programs
	Plugins     *pluginsreg.Registry
	GetSecret   func(context.Context, apiproject.ProjectID, string) (string, error)
	GetCreds    func(context.Context, apiproject.ProjectID, string, string) ([]byte, error)
	StateStore  statestore.Store
	L           L.Nullable

	worker worker.Worker
}

func (s *Sessions) Init() {
	s.worker = worker.New(s.Temporal, "sessions", worker.Options{})
	s.worker.RegisterWorkflow(s.signalWaitingSessions)
}

func (s *Sessions) Start() error { return s.worker.Start() }

// First returned module is main.
// TODO: account for internal modules.
func (s *Sessions) loadAllModules(
	ctx workflow.Context,
	pid apiproject.ProjectID,
	predecls []string,
	mainPath *apiprogram.Path,
) ([]*apiprogram.Module, []apiplugin.PluginID, error) {
	l := s.L.With("project_id", pid, "main", mainPath.String())

	const N = 16

	q := make([]*apiprogram.Path, 0, N)

	mods := make([]*apiprogram.Module, 0, N)
	pluginIDs := make([]apiplugin.PluginID, 0, N)

	load := func(curr *apiprogram.Path, depth int) error {
		l := l.With("path", curr.String())

		var mod *apiprogram.Module

		l.Debug("loading")

		// first is internal boot. second is the actual main.
		if curr.String() == "$inmem:second" {
			curr = mainPath
		}

		if err := workflow.ExecuteLocalActivity(
			// TODO: this assumes all failures are deterministic.
			withLocalActivityWithoutRetries(ctx),
			s.Programs.Load,
			pid,
			predecls,
			curr,
		).Get(ctx, &mod); err != nil {
			temporalErrorLogger(l, err)("load error")
			return err
		}

		mods = append(mods, mod)

		var depPaths []*apiprogram.Path

		if err := workflow.ExecuteLocalActivity(ctx, langtools.GetModuleDependencies, s.Programs.Catalog, mod).Get(ctx, &depPaths); err != nil {
			err = fmt.Errorf("dependencies: %w", err)

			temporalErrorLogger(l, err)("dependencies error")
			l.Debug("dependencies error", "err", err)
			return err
		}

		l.Debug("dependencies", "deps", depPaths)

		for _, depPath := range depPaths {
			if depPath.IsInternal() {
				if !curr.IsInternal() {
					return fmt.Errorf("internal loads are allowed only from internal modules")
				}
			} else if plugID, isPlugin := depPath.PluginID(); isPlugin {
				// handled by either the lang load or session load (plugin).
				pluginIDs = append(pluginIDs, plugID)
				continue
			}

			q = append(q, depPath)
		}

		return nil
	}

	if err := load(apiprogram.MustParsePathString("$internal:boot.kitteh"), 0); err != nil {
		return nil, nil, err
	}

	for depth := 1; len(q) != 0; q, depth = q[1:], depth+1 {
		curr := q[0]

		if err := load(curr, depth); err != nil {
			return nil, nil, err
		}
	}

	return mods, pluginIDs, nil
}

func (s *Sessions) signalWaitingSessions(ctx workflow.Context, event *apievent.Event, project *apiproject.Project, binding string) error {
	l := s.L.With("event_id", event.ID(), "event_type", event.Type(), "project_id", project.ID())

	ctx = workflow.WithLocalActivityOptions(
		ctx,
		workflow.LocalActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
		},
	)

	var eids []apievent.EventID

	// TODO: get only sessions that wait on a specific event name.
	if err := workflow.ExecuteLocalActivity(ctx, s.EventsStore.GetProjectWaitingEvents, project.ID()).Get(ctx, &eids); err != nil {
		l.Error("get waiting error", "err", err)
		return err
	}

	l.Debug("waiting sessions", "event_ids", eids)

	data := akmod.NewEventSignal(binding, event)

	var futs []workflow.Future

	for _, eid := range eids {
		l := l.With("target_event_id", eid)

		l.Debug("signaling waiting session")

		workflowID := events.GetIngestProjectEventWorkflowID(eid, project.ID())

		fut := workflow.ExecuteLocalActivity(
			ctx,
			func(ctx context.Context, wid string, data interface{}) error {
				if err := s.Temporal.SignalWorkflow(ctx, wid, "", akmod.SessionEventSignalName, data); err != nil {
					l := s.L.With("workflow_id", workflowID)
					if _, ok := err.(*serviceerror.NotFound); ok { // for some reason errors.Is doesn't work well here.
						l.Debug("workflow not found, might have already finished")
						return nil
					}

					l.Error("signal error", "err", err)
				}

				return nil
			},
			workflowID,
			data,
		)

		futs = append(futs, fut)
	}

	for _, f := range futs {
		if err := f.Get(ctx, nil); err != nil {
			l.Error("signal error", "err", err)
		}
	}

	return nil
}

// TODO: this should probably run as a child-workflow.
func (s *Sessions) Run(
	ctx workflow.Context,
	event *apievent.Event,
	project *apiproject.Project,
	srcBindingName string,
) (*apilang.RunSummary, error) {
	l := s.L.With("event_id", event.ID(), "project_id", project.ID(), "event_source_binding_name", srcBindingName)

	sessionID := fmt.Sprintf("%v/%v", project.ID(), event.ID())

	if err := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(
			ctx,
			workflow.ChildWorkflowOptions{
				TaskQueue:         "sessions",
				ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
			},
		),
		s.signalWaitingSessions,
		event,
		project,
		srcBindingName,
	).Get(ctx, nil); err != nil { // TODO: This need to be fire and forget. For some reason, if no get is done the workflow will not run.
		l.Error("signal workflow error", "err", err)
	}

	mainPath := project.Settings().MainPath()

	if mainPath.IsInternal() {
		return nil, fmt.Errorf("invalid main path %q", mainPath.String())
	}

	predecls := project.Settings().Predecls()

	predeclsKeys := make([]string, 0, len(predecls))
	for k := range predecls {
		predeclsKeys = append(predeclsKeys, k)
	}

	l.Debug("loading all modules", "predecls", predeclsKeys)

	mods, depPluginIDs, err := s.loadAllModules(
		ctx,
		project.ID(),
		predeclsKeys,
		mainPath,
	)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	akmodPlugin := akmod.New(
		s.L.Named("akmod"),
		s.StateStore,
		project,
		event,
		func(ctx context.Context, n string) (string, error) {
			// no need for an activity - this is part of a the Call activity.
			return s.GetSecret(ctx, project.ID(), n)
		},
		func(ctx context.Context, k, n string) ([]byte, error) {
			// no need for an activity - this is part of a the Call activity.
			return s.GetCreds(ctx, project.ID(), k, n)
		},
		srcBindingName,
		"0.0.0", /* TODO */
	)

	plugins := map[apiplugin.PluginID]plugin.Plugin{
		akmod.PluginID: akmodPlugin,
	}

	for _, id := range depPluginIDs {
		l := l.With("plugin_id", id)

		var projectPlugin *apiproject.ProjectPlugin

		if plugins[id] != nil {
			// already registered, no need to repeat.
			continue
		} else if id.IsInternal() {
			// implemented internaly by ak.
		} else if projectPlugin = project.Settings().Plugin(id); projectPlugin != nil {
			if !projectPlugin.Enabled() {
				return nil, fmt.Errorf("plugin %q is not enabled", id)
			}
		} else {
			// might be builtin in starlark (or lang specific).
			return nil, fmt.Errorf("plugin not configured for project")
		}

		var pl plugin.Plugin

		// TODO: this might block for a short while, so better put it into some kind of activity?
		pl, err := s.Plugins.NewPlugin(context.Background() /* TODO */, l.Named("plugin:"+id.String()), id, sessionID)
		if err != nil {
			return nil, fmt.Errorf("cannot create plugin %v: %w", id, err)
		}

		plugins[id] = pl
	}

	var prints []string

	evSigCh := workflow.GetSignalChannel(ctx, akmod.SessionEventSignalName)

	sessionContext := &akmod.Session{
		L:              l.Named("signals"),
		Context:        ctx,
		Event:          event,
		SignalChannel:  evSigCh,
		Temporal:       s.Temporal,
		ProjectID:      project.ID(),
		SrcBindingName: srcBindingName,
		UpdateState: func(state *apievent.ProjectEventState) error {
			return s.updateProjectState(ctx, event.ID(), project.ID(), state)
		},
	}

	runCtx := akmod.WithSessionContext(
		// TODO: should context.Background() be replaced with something else to affect cancellations
		//       or should we just rely on underlying activity cancellation?
		context.Background(),
		sessionContext,
	)

	runEnv := lang.RunEnv{
		Scope: fmt.Sprintf("%v_%v", event.ID(), project.ID()),
		Print: func(s string) {
			l.Debug("print", "text", s)
			prints = append(prints, s)
		},
		Predecls: predecls,
		Load: func(_ context.Context, path *apiprogram.Path) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
			l := l.With("path", path.String())

			plugID, isPlugin := path.PluginID()

			// Only plugins can be loaded here as all sources are already included via loadAllModules.
			if !isPlugin {
				return nil, nil, L.Error(l, "load not possible, not a plugin")
			}

			l = l.With("plugin_id", plugID)

			l.Debug("loading plugin")

			plug := plugins[plugID]
			if plug == nil {
				return nil, nil, L.Error(l, "plugin not preloaded")
			}

			var members map[string]*apivalues.Value

			if err := workflow.ExecuteLocalActivity(ctx, plug.GetAll).Get(ctx, &members); err != nil {
				return nil, nil, fmt.Errorf("load members: %w", err)
			}

			for _, m := range members {
				if err := apivalues.Walk(m, func(curr, _ *apivalues.Value, _ apivalues.Role) error {
					if _, ok := curr.Get().(apivalues.CallValue); ok {
						if err := apivalues.SetCallIssuer(curr, plugID.String()); err != nil {
							return L.Error(l, "set call issuer error", "err", err)
						}
					}

					return nil
				}); err != nil {
					return nil, nil, err
				}
			}

			if plugName := plugID.PluginName().String(); members[plugName] == nil {
				members[plugName] = apivalues.Module(plugName, members)
			}

			return members, nil, nil
		},
		Call: func(runCtx context.Context, callv *apivalues.Value, kwargs map[string]*apivalues.Value, args []*apivalues.Value, sum *apilang.RunSummary) (*apivalues.Value, error) {
			l := l.With("call", callv.String())

			sessionContext.RunSummary = sum

			callvCall := apivalues.GetCallValue(callv)
			if callvCall == nil {
				return nil, L.Error(l, "call to non-call value")
			}

			plug, ok := plugins[apiplugin.PluginID(callvCall.Issuer)]
			if !ok {
				return nil, L.Error(l, "call to unknown plugin", "call", callvCall)
			}

			allowPassingCallValues := callvCall.Flags["allow_passing_call_values"]

			if !allowPassingCallValues {
				// These ensures that any call value returned from a plugin call is generated by that call.
				// This is essential for proper call registeration as it prevent plugins mascarading as
				// other plugins in return vals.
				if err := EnsureNoCallValuesInArgs(args); err != nil {
					return nil, err
				}

				if err := EnsureNoCallValuesInKWArgs(kwargs); err != nil {
					return nil, err
				}
			}

			call := func(ctx context.Context, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
				l := l.With("name", callvCall.Name)

				l.Debug("invoking", "args", args, "kwargs", kwargs)

				v, err := plug.Call(ctx, callv, args, kwargs)

				l.Debug("returned", "err", err, "v", v)

				if v != nil {
					if err := apivalues.Walk(v, func(curr, _ *apivalues.Value, _ apivalues.Role) error {
						if currcv, ok := curr.Get().(apivalues.CallValue); ok {
							if currcv.Issuer == "" {
								if err := apivalues.SetCallIssuer(curr, callvCall.Issuer); err != nil {
									return L.Error(l, "set call issuer error", "err", err)
								}
							} else if !allowPassingCallValues && callvCall.Issuer != currcv.Issuer {
								// don't let the plugin fool the session into calling another plugin's call value.
								return L.Error(l, "invalid issuer returned by call", "returned", currcv.Issuer, "expected", callvCall.Issuer)
							}
						}

						return nil
					}); err != nil {
						return nil, err
					}
				}

				return v, err
			}

			var ret *apivalues.Value

			if callvCall.Flags["session"] {
				// Run directly in workflow context.

				// callCtx is the context given to RunModule.
				ret, err = call(runCtx, kwargs)
			} else {
				err = workflow.ExecuteLocalActivity(
					// TODO: this assume call has no non-deterministic failures.
					withLocalActivityWithoutRetries(ctx),
					call,
					kwargs,
				).Get(ctx, &ret)
			}

			if err != nil {
				return nil, fmt.Errorf("%v: %w", callvCall, err)
			}

			return ret, nil
		},
	}

	// This should not be in an activity as this is determinisitc. Any non-deterministic
	// action is being run as an activity in runEnv.Call above.
	_, sum, err := langtools.RunModules(runCtx, s.Programs.Catalog, &runEnv, mods)
	if err != nil {
		return sum, fmt.Errorf("run: %w", err)
	}

	l.Debug("run completed", "summary", sum)

	// TODO: this should happen even if session fails.
	s.Plugins.CloseSession(sessionID)

	return sum, nil
}

func (s *Sessions) updateProjectState(ctx workflow.Context, id apievent.EventID, pid apiproject.ProjectID, state *apievent.ProjectEventState) error {
	l := s.L.With("event_id", id, "project_id", pid, "state", state)

	l.Debug("updating project state")

	if err := workflow.ExecuteLocalActivity(ctx, s.EventsStore.UpdateStateForProject, id, pid, state).Get(ctx, nil); err != nil {
		l.Error("update project event state failed", "err", err)
		return err
	}

	return nil
}
