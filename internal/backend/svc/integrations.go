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
	"go.autokitteh.dev/autokitteh/integrations/redis"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
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

func integrationsFXOption() fx.Option {
	return fx.Options(
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
		integration("redis", configset.Empty, redis.New),
		integration("sheets", configset.Empty, sheets.New),
		integration("slack", configset.Empty, slack.New),
		integration("twilio", configset.Empty, twilio.New),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, svcs sdkservices.Services) {
			HookOnStart(lc, func(ctx context.Context) error {
				asana.Start(l, muxes)
				auth0.Start(l, muxes, svcs.Vars())
				aws.Start(l, muxes)
				chatgpt.Start(l, muxes)
				confluence.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
				discord.Start(l, muxes, svcs.Vars(), svcs.Dispatcher())
				github.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
				gemini.Start(l, muxes)
				google.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
				hubspot.Start(l, svcs.OAuth(), muxes)
				jira.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
				slack.Start(l, muxes, svcs.Vars(), svcs.Dispatcher())
				twilio.Start(l, muxes, svcs.Vars(), svcs.Dispatcher())
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
