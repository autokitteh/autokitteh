package svc

import (
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type runtimesConfig struct{}

var runtimesConfigs = configset.Set[runtimesConfig]{
	Default: &runtimesConfig{},
}

func runtime[T any](name string, cfg configset.Set[T], init any) fx.Option {
	return Component(name, cfg, fx.Provide(
		fx.Annotate(
			init,
			fx.ResultTags(`group:"runtimes"`),
		),
	))
}

func runtimesFXOption() fx.Option {
	return fx.Options(
		runtime("starlarkrt", starlarkrt.Configs, starlarkrt.NewFromConfig),
		runtime("configrt", configset.Empty, configrt.New),
		runtime("pythonrt", configset.Empty, pythonrt.New),

		Component(
			"runtimes",
			runtimesConfigs,
			fx.Provide(
				fx.Annotate(
					func(rts []*sdkruntimes.Runtime, cfg *runtimesConfig) sdkservices.Runtimes {
						return kittehs.Must1(sdkruntimes.New(rts))
					},
					fx.ParamTags(`group:"runtimes"`),
				),
			),
		),
	)
}
