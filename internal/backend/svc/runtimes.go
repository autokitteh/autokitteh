package svc

import (
	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt"
	//"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type runtimesConfig struct{}

var runtimesConfigs = configset.Set[runtimesConfig]{
	Default: &runtimesConfig{},
}

func runtime[T any](name string, cfg configset.Set[T], init any, opts ...fx.Option) fx.Option {
	opts = append(
		[]fx.Option{
			fx.Provide(
				fx.Annotate(
					init,
					fx.ResultTags(`group:"runtimes"`),
				),
			),
		},
		opts...,
	)

	return Component(name, cfg, opts...)
}

func runtimesFXOption() fx.Option {
	return fx.Options(
		runtime("starlarkrt", configset.Empty, starlarkrt.New),
		runtime("configrt", configset.Empty, configrt.New),
		runtime(
			"nodejsrt",
			nodejsrt.Configs,
			func(cfg *nodejsrt.Config, l *zap.Logger, httpsvc httpsvc.Svc) (*sdkruntimes.Runtime, error) {
				return nodejsrt.New(cfg, l, httpsvc.Addr)
			},
			fx.Invoke(func(l *zap.Logger, muxes *muxes.Muxes) {
				nodejsrt.ConfigureWorkerGRPCHandler(l, muxes.NoAuth)
			}),
		),
		//runtime(
		//	"pythonrt",
		//	pythonrt.Configs,
		//	func(cfg *pythonrt.Config, l *zap.Logger, httpsvc httpsvc.Svc) (*sdkruntimes.Runtime, error) {
		//		return pythonrt.New(cfg, l, httpsvc.Addr)
		//	},
		//	fx.Invoke(func(l *zap.Logger, muxes *muxes.Muxes) {
		//		pythonrt.ConfigureWorkerGRPCHandler(l, muxes.NoAuth)
		//	}),
		//),

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
