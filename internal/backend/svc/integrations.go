package svc

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integrationsConfig struct {
	Test bool `koanf:"test"`
}

var integrationConfigs = configset.Set[integrationsConfig]{
	Default: &integrationsConfig{},
	Dev:     &integrationsConfig{Test: true},
}

type sysVars struct{ vs sdkservices.Vars }

func (vs sysVars) Set(ctx context.Context, v ...sdktypes.Var) error {
	return vs.vs.Set(authcontext.SetAuthnSystemUser(ctx), v...)
}

func (vs sysVars) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
	return vs.vs.Delete(authcontext.SetAuthnSystemUser(ctx), sid, names...)
}

func (vs sysVars) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	return vs.vs.Get(authcontext.SetAuthnSystemUser(ctx), sid, names...)
}

func (vs sysVars) FindActiveConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return vs.vs.FindActiveConnectionIDs(authcontext.SetAuthnSystemUser(ctx), iid, name, value)
}

func integrationsFXOption() fx.Option {
	inits := fx.Options(kittehs.Transform(integrations.All(), func(i integrations.Integration) fx.Option {
		return Component(i.Name, configset.Empty, fx.Provide(
			fx.Annotate(i.Init, fx.ResultTags(`group:"integrations"`)),
		))
	})...)

	return fx.Module(
		"integrations",

		fx.Decorate(func(vs sdkservices.Vars) sdkservices.Vars { return sysVars{vs} }),
		fx.Decorate(func(dispatch sdkservices.DispatchFunc) sdkservices.DispatchFunc {
			return func(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (*sdkservices.DispatchResponse, error) {
				return dispatch(authcontext.SetAuthnSystemUser(ctx), event, opts)
			}
		}),

		inits,

		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, vars sdkservices.Vars, oauth *oauth.OAuth, dispatch sdkservices.DispatchFunc) {
			l.Info("supported integrations", zap.Strings("integrations", integrations.Names()))

			HookOnStart(lc, func(ctx context.Context) error {
				for _, i := range integrations.All() {
					if i.Start != nil {
						l.Debug("starting integration", zap.String("integration", i.Name))

						i.Start(l, muxes, vars, oauth, dispatch)
					}
				}

				return nil
			})
		}),
		Component(
			"integrations",
			integrationConfigs,
			fx.Provide(
				fx.Annotate(
					func(is []sdkservices.Integration, cfg *integrationsConfig, vars sdkservices.Vars) sdkservices.Integrations {
						if cfg.Test {
							is = append(is, integrations.NewTestIntegration(vars))
						}

						return sdkintegrations.New(is)
					},
					fx.ParamTags(`group:"integrations"`),
				),
			),
		),
	)
}
