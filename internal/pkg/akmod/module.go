package akmod

import (
	"context"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/plugin/builtinplugin"
	"go.autokitteh.dev/sdk/pluginimpl"

	"github.com/autokitteh/autokitteh/internal/pkg/statestore"

	"github.com/autokitteh/L"
)

var PluginID = apiplugin.NewInternalPluginID("ak")

func New(
	l L.L,
	stateStore statestore.Store,
	project *apiproject.Project,
	event *apievent.Event,
	getSecret_ func(context.Context, string) (string, error),
	getCred_ func(context.Context, string, string) ([]byte, error),
	bindingName string,
	version string,
) *builtinplugin.BuiltinPlugin {
	l = L.N(l)

	return &builtinplugin.BuiltinPlugin{
		Plugin: &pluginimpl.Plugin{
			Doc: "builtin autokitteh module",
			Members: map[string]*pluginimpl.PluginMember{
				"version":      pluginimpl.NewValueMember("ak version", apivalues.String(version)),
				"event_source": pluginimpl.NewValueMember("event source binding name", apivalues.String(bindingName)),
				"event":        pluginimpl.NewValueMember("event", event.AsValue()),
				"nop": pluginimpl.NewSimpleMethodMember(
					"no-op",
					func(context.Context, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, error) {
						return apivalues.None, nil
					},
				),
				"get_secret":      pluginimpl.NewSimpleMethodMember("get preset secret", getSecret(getSecret_)),
				"get_credentials": pluginimpl.NewSimpleMethodMember("get preset credentials", getCreds(getCred_)),
				"state": pluginimpl.NewLazyValueMember("state storage", (&state{
					projectID:  project.ID(),
					stateStore: stateStore,
					l:          l.Named("state"),
				}).asValue),
				"sources": pluginimpl.NewLazyValueMember("event sources control", (&sources{
					l:              l.Named("sources"),
					srcBindingName: bindingName,
					event:          event,
					bindings:       &bindings{},
				}).asValueWithMatch),
				"signals": pluginimpl.NewLazyValueMember("signals control", (&signals{
					l: l.Named("signals"),
				}).asValue),
			},
		},
	}
}
