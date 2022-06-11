package sessions

import (
	"context"
	"fmt"

	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/plugin"

	"github.com/autokitteh/L"

	"github.com/autokitteh/autokitteh/internal/pkg/akmod"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/pluginsreg"
	"github.com/autokitteh/autokitteh/internal/pkg/programs"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
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
}

func (s *Sessions) Start() error { return s.worker.Start() }

// TODO: this should probably run as a child-workflow.
func (s *Sessions) Run(
	ctx workflow.Context,
	event *apievent.Event,
	project *apiproject.Project,
	srcBindingName string,
) (*apilang.RunSummary, error) {
	l := s.L.With("event_id", event.ID(), "project_id", project.ID(), "event_source_binding_name", srcBindingName)

	sessionID := fmt.Sprintf("%v/%v", project.ID(), event.ID())

	mainPath := project.Settings().MainPath()

	if mainPath.IsInternal() {
		return nil, fmt.Errorf("invalid main path %q", mainPath.String())
	}

	predecls := project.Settings().Predecls()

	predeclsKeys := make([]string, 0, len(predecls))
	for k := range predecls {
		predeclsKeys = append(predeclsKeys, k)
	}

	l.Debug("fetching all modules", "predecls", predeclsKeys)

	if err := s.updateProjectState(ctx, event.ID(), project.ID(), apievent.NewLoadingProjectEventState(mainPath)); err != nil {
		return nil, err
	}

	var fr programs.FetchResult

	if err := workflow.ExecuteLocalActivity(
		ctx,
		s.Programs.Fetch,
		project.ID(),
		mainPath,
		predeclsKeys,
	).Get(ctx, &fr); err != nil {
		err = fmt.Errorf("fetch: %w", err)

		temporalErrorLogger(l, err)("fetch error")
		l.Debug("fetch error", "err", err)
		return nil, err
	}

	if err := s.updateProjectState(ctx, event.ID(), project.ID(), apievent.NewLoadedProjectEventState(fr.Paths())); err != nil {
		return nil, err
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

	for _, id := range fr.PluginIDs {
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

			var (
				ret *apivalues.Value
				err error
			)

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

	if err := s.updateProjectState(ctx, event.ID(), project.ID(), apievent.NewRunningProjectEventState()); err != nil {
		return nil, err
	}

	// This should not be in an activity as this is determinisitc. Any non-deterministic
	// action is being run as an activity in runEnv.Call above.
	_, sum, err := langtools.RunModules(runCtx, s.Programs.Catalog, &runEnv, fr.Modules())
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
