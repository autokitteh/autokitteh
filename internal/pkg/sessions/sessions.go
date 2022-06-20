package sessions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
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

type Config struct {
	TaskQueueName        string        `envconfig:"TASK_QUEUE_NAME" default:"sessions" json:"task_queue_name"`
	UpdateStateTimeout   time.Duration `envconfig:"UPDATE_STATE_TIMEOUT" default:"30s" json:"update_state_timeout"`
	ProgramsFetchTimeout time.Duration `envconfig:"PROGRAMS_FETCH_TIMEOUT" default:"1m" json:"programs_fetch_timeout"`
	LoadPluginsTimeout   time.Duration `envconfig:"LOAD_PLUGINS_TIMEOUT" default:"1m" json:"load_plugins_timeout"`
}

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

	// Since all activities run in the same temporal session,
	// they will run in the same process. This enables them to
	// share the plugins value.
	pluginsMutex sync.RWMutex
	plugins      map[string]map[apiplugin.PluginID]plugin.Plugin
}

const (
	programsFetchActivityName         = "programs-fetch"
	updateStateForProjectActivityName = "update-state-for-project"
	loadPluginsActivityName           = "load-plugins"
	loadPluginActivityName            = "load-plugin"
	callPluginActivityName            = "call-plugin"
)

func (s *Sessions) Init() {
	s.worker = worker.New(
		s.Temporal,
		s.Config.TaskQueueName,
		worker.Options{
			EnableSessionWorker: true,
		},
	)

	s.worker.RegisterActivityWithOptions(
		s.Programs.Fetch,
		activity.RegisterOptions{Name: programsFetchActivityName},
	)

	s.worker.RegisterActivityWithOptions(
		s.EventsStore.UpdateStateForProject,
		activity.RegisterOptions{Name: updateStateForProjectActivityName},
	)

	s.worker.RegisterActivityWithOptions(
		s.loadPlugins,
		activity.RegisterOptions{Name: loadPluginsActivityName},
	)

	s.worker.RegisterActivityWithOptions(
		s.loadPlugin,
		activity.RegisterOptions{Name: loadPluginActivityName},
	)

	s.worker.RegisterActivityWithOptions(
		s.callPlugin,
		activity.RegisterOptions{Name: callPluginActivityName},
	)
}

func (s *Sessions) Start() error { return s.worker.Start() }

// TODO: this should probably run as a child-workflow.
func (s *Sessions) Run(
	ctx workflow.Context,
	event *apievent.Event,
	project *apiproject.Project,
	srcBindingName string,
) (*apilang.RunSummary, error) {
	ctx = workflow.WithTaskQueue(ctx, s.Config.TaskQueueName)

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

	if err := workflow.ExecuteActivity(
		workflow.WithStartToCloseTimeout(ctx, s.Config.ProgramsFetchTimeout),
		programsFetchActivityName,
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

	sessionCtx, err := workflow.CreateSession(
		ctx,
		&workflow.SessionOptions{
			ExecutionTimeout: time.Hour, // this probably needs to be configurable.
			CreationTimeout:  time.Minute,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	defer workflow.CompleteSession(sessionCtx)

	if err := workflow.ExecuteActivity(
		workflow.WithStartToCloseTimeout(sessionCtx, s.Config.LoadPluginsTimeout),
		loadPluginsActivityName,
		sessionID,
		project,
		event,
		srcBindingName,
		fr,
	).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("load plugins: %w", err)
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

			var members map[string]*apivalues.Value

			// This needs to be an acitivity to guarntee it's part of the session.
			if err := workflow.ExecuteActivity(
				workflow.WithStartToCloseTimeout(sessionCtx, 5*time.Second),
				loadPluginActivityName,
				sessionID,
				plugID,
			).Get(ctx, &members); err != nil {
				return nil, nil, L.Error(l, "load plugins: %w", err)
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

			if !callvCall.Flags["allow_passing_call_values"] { // [# allow_passing_call_values #]
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

			var (
				ret *apivalues.Value
				err error
			)

			if callvCall.Flags["session"] {
				// Run directly in workflow context.

				// callCtx is the context given to RunModule.
				ret, err = s.callPlugin(runCtx, sessionID, callv, args, kwargs)
			} else {
				err = workflow.ExecuteActivity(
					workflow.WithRetryPolicy(
						workflow.WithStartToCloseTimeout(sessionCtx, 5*time.Second),
						temporal.RetryPolicy{ // TODO
							MaximumAttempts: 1,
						},
					),
					callPluginActivityName,
					sessionID,
					callv,
					args,
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

	s.pluginsMutex.Lock()
	delete(s.plugins, sessionID)
	s.pluginsMutex.Unlock()

	// TODO: this should happen even if session fails.
	s.Plugins.CloseSession(sessionID)

	return sum, nil
}

func (s *Sessions) updateProjectState(ctx workflow.Context, id apievent.EventID, pid apiproject.ProjectID, state *apievent.ProjectEventState) error {
	l := s.L.With("event_id", id, "project_id", pid, "state", state.Name())

	l.Debug("updating project state")

	if err := workflow.ExecuteActivity(
		workflow.WithStartToCloseTimeout(ctx, s.Config.UpdateStateTimeout),
		updateStateForProjectActivityName,
		id,
		pid,
		state,
	).Get(ctx, nil); err != nil {
		l.Error("update project event state failed", "err", err)
		return err
	}

	return nil
}
