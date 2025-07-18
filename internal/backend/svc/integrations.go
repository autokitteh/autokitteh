package svc

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"logur.dev/logur/integration/grpc"

	"go.autokitteh.dev/autokitteh/integrations/airtable"
	"go.autokitteh.dev/autokitteh/integrations/anthropic"
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
	"go.autokitteh.dev/autokitteh/integrations/height"
	"go.autokitteh.dev/autokitteh/integrations/hubspot"
	"go.autokitteh.dev/autokitteh/integrations/kubernetes"
	"go.autokitteh.dev/autokitteh/integrations/linear"
	"go.autokitteh.dev/autokitteh/integrations/microsoft"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/teams"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/integrations/reddit"
	"go.autokitteh.dev/autokitteh/integrations/salesforce"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/integrations/zoom"
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

		integration("airtable", configset.Empty, airtable.New),
		integration("anthropic", configset.Empty, anthropic.New),
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
		integration("height", configset.Empty, height.New),
		integration("hubspot", configset.Empty, hubspot.New),
		integration("jira", configset.Empty, jira.New),
		integration("kubernetes", configset.Empty, kubernetes.New),
		integration("linear", configset.Empty, linear.New),
		integration("microsoft", configset.Empty, microsoft.New),
		integration("microsoft_teams", configset.Empty, teams.New),
		integration("reddit", configset.Empty, reddit.New),
		integration("salesforce", configset.Empty, salesforce.New),
		integration("sheets", configset.Empty, sheets.New),
		integration("slack", configset.Empty, slack.New),
		integration("twilio", configset.Empty, twilio.New),
		integration("zoom", configset.Empty, zoom.New),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, vars sdkservices.Vars, oauth *oauth.OAuth, dispatch sdkservices.DispatchFunc) {
			HookOnStart(lc, func(ctx context.Context) error {
				airtable.Start(l, muxes, vars, oauth, dispatch)
				anthropic.Start(l, muxes, vars)
				asana.Start(l, muxes)
				auth0.Start(l, muxes, vars)
				aws.Start(l, muxes)
				chatgpt.Start(l, muxes)
				confluence.Start(l, muxes, vars, oauth, dispatch)
				discord.Start(l, muxes, vars, dispatch)
				gemini.Start(l, muxes)
				github.Start(l, muxes, vars, oauth, dispatch)
				google.Start(l, muxes, vars, oauth, dispatch)
				height.Start(l, muxes, vars, oauth, dispatch)
				hubspot.Start(l, muxes, oauth)
				jira.Start(l, muxes, vars, oauth, dispatch)
				kubernetes.Start(l, muxes)
				linear.Start(l, muxes, vars, oauth, dispatch)
				microsoft.Start(l, muxes, vars, oauth, dispatch)
				reddit.Start(l, muxes, vars)
				salesforce.Start(l, muxes, vars, oauth, dispatch)
				slack.Start(l, muxes, vars, dispatch)
				twilio.Start(l, muxes, vars, dispatch)
				zoom.Start(l, muxes, vars, oauth, dispatch)
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
