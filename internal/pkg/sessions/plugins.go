package sessions

import (
	"context"
	"fmt"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/plugin"

	"github.com/autokitteh/L"

	"github.com/autokitteh/autokitteh/internal/pkg/akmod"
	"github.com/autokitteh/autokitteh/internal/pkg/programs"
)

func (s *Sessions) loadPlugins(ctx context.Context, sessionID string, project *apiproject.Project, event *apievent.Event, srcBindingName string, fr *programs.FetchResult) error {
	l := s.L.Named("loadplugins")

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
				return fmt.Errorf("plugin %q is not enabled", id)
			}
		} else {
			// might be builtin in starlark (or lang specific).
			return fmt.Errorf("plugin not configured for project")
		}

		var pl plugin.Plugin

		// TODO: this might block for a short while, so better put it into some kind of activity?
		pl, err := s.Plugins.NewPlugin(context.Background() /* TODO */, l.Named("plugin:"+id.String()), id, sessionID)
		if err != nil {
			return fmt.Errorf("cannot create plugin %v: %w", id, err)
		}

		plugins[id] = pl
	}

	s.pluginsMutex.Lock()

	if s.plugins == nil {
		s.plugins = make(map[string]map[apiplugin.PluginID]plugin.Plugin)
	}

	s.plugins[sessionID] = plugins

	s.pluginsMutex.Unlock()

	return nil
}

func (s *Sessions) loadPlugin(ctx context.Context, sessionID string, plugID apiplugin.PluginID) (map[string]*apivalues.Value, error) {
	l := s.L.With("plugin_id", plugID)

	l.Debug("loading plugin")

	s.pluginsMutex.RLock()

	plugins := s.plugins[sessionID]
	if plugins == nil {
		s.pluginsMutex.RUnlock()
		return nil, fmt.Errorf("session not found")
	}

	s.pluginsMutex.RUnlock()

	plug := plugins[plugID]
	if plug == nil {
		return nil, L.Error(l, "plugin not preloaded")
	}

	return plug.GetAll(ctx)
}

func (s *Sessions) callPlugin(
	ctx context.Context,
	sessionID string,
	callv *apivalues.Value,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	l := s.L.With("call", callv)

	callvCall := apivalues.GetCallValue(callv)
	if callvCall == nil {
		return nil, L.Error(l, "call to non-call value")
	}

	s.pluginsMutex.RLock()
	plugins := s.plugins[sessionID]
	s.pluginsMutex.RUnlock()

	if plugins == nil {
		return nil, L.Error(l, "session not found")
	}

	plug, ok := plugins[apiplugin.PluginID(callvCall.Issuer)]
	if !ok {
		return nil, L.Error(l, "call to unknown plugin", "call", callvCall)
	}

	l.Debug("invoking", "name", callvCall.Name, "args", args, "kwargs", kwargs)

	v, err := plug.Call(ctx, callv, args, kwargs)

	l.Debug("returned", "err", err, "v", v)

	if v != nil {
		if err := apivalues.Walk(v, func(curr, _ *apivalues.Value, _ apivalues.Role) error {
			if currcv, ok := curr.Get().(apivalues.CallValue); ok {
				if currcv.Issuer == "" {
					if err := apivalues.SetCallIssuer(curr, callvCall.Issuer); err != nil {
						return L.Error(l, "set call issuer error", "err", err)
					}
				} else if !callvCall.Flags["allow_passing_call_values"] && callvCall.Issuer != currcv.Issuer { // [# allow_passing_call_values #]
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
