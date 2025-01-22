package svc

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/asana"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/confluence"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/jira"
	"go.autokitteh.dev/autokitteh/integrations/auth0"
	"go.autokitteh.dev/autokitteh/integrations/aws"
	"go.autokitteh.dev/autokitteh/integrations/chatgpt"
	"go.autokitteh.dev/autokitteh/integrations/discord"
	"go.autokitteh.dev/autokitteh/integrations/github"
	"go.autokitteh.dev/autokitteh/integrations/google"
	"go.autokitteh.dev/autokitteh/integrations/google/calendar"
	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gemini"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/sheets"
	"go.autokitteh.dev/autokitteh/integrations/grpc"
	"go.autokitteh.dev/autokitteh/integrations/hubspot"
	"go.autokitteh.dev/autokitteh/integrations/microsoft"
	"go.autokitteh.dev/autokitteh/integrations/redis"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
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

func integration[T any](name string, cfg configset.Set[T], init any) fx.Option {
	return Component(name, cfg, fx.Provide(
		fx.Annotate(
			init,
			fx.ResultTags(`group:"integrations"`),
		),
	))
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

func (vs sysVars) FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return vs.vs.FindConnectionIDs(authcontext.SetAuthnSystemUser(ctx), iid, name, value)
}

func integrationsFXOption() fx.Option {
	return fx.Module(
		"integrations",

		fx.Decorate(func(vs sdkservices.Vars) sdkservices.Vars { return sysVars{vs} }),
		fx.Decorate(func(dispatch sdkservices.DispatchFunc) sdkservices.DispatchFunc {
			return func(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
				return dispatch(authcontext.SetAuthnSystemUser(ctx), event, opts)
			}
		}),

		integration("asana", configset.Empty, asana.New),
		integration("auth0", configset.Empty, auth0.New),
		integration("aws", configset.Empty, aws.New),
		integration("calendar", configset.Empty, calendar.New),
		integration("chatgpt", configset.Empty, chatgpt.New),
		integration("confluence", configset.Empty, confluence.New),
		integration("discord", configset.Empty, discord.New),
		integration("drive", configset.Empty, drive.New),
		integration("forms", configset.Empty, forms.New),
		integration("github", configset.Empty, github.New),
		integration("gmail", configset.Empty, gmail.New),
		integration("gemini", configset.Empty, gemini.New),
		integration("google", configset.Empty, google.New),
		integration("grpc", configset.Empty, grpc.New),
		integration("hubspot", configset.Empty, hubspot.New),
		integration("jira", configset.Empty, jira.New),
		integration("microsoft", configset.Empty, microsoft.New),
		integration("redis", configset.Empty, redis.New),
		integration("sheets", configset.Empty, sheets.New),
		integration("slack", configset.Empty, slack.New),
		integration("twilio", configset.Empty, twilio.New),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, vars sdkservices.Vars, dispatch sdkservices.DispatchFunc, oauth sdkservices.OAuth) {
			HookOnStart(lc, func(ctx context.Context) error {
				asana.Start(l, muxes)
				auth0.Start(l, muxes, vars)
				aws.Start(l, muxes)
				chatgpt.Start(l, muxes)
				confluence.Start(l, muxes, vars, oauth, dispatch)
				discord.Start(l, muxes, vars, dispatch)
				gemini.Start(l, muxes)
				github.Start(l, muxes, vars, oauth, dispatch)
				google.Start(l, muxes, vars, oauth, dispatch)
				hubspot.Start(l, muxes, oauth)
				jira.Start(l, muxes, vars, oauth, dispatch)
				microsoft.Start(l, muxes, vars, oauth, dispatch)
				slack.Start(l, muxes, vars, dispatch)
				twilio.Start(l, muxes, vars, dispatch)
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
