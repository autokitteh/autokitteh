package svc

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/asana"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/confluence"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/jira"
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
	"go.autokitteh.dev/autokitteh/integrations/redis"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type integrationsConfig struct {
	Test bool `koanf:"test"`
}

func (integrationsConfig) Validate() error { return nil }

var integrationConfigs = config.Set[integrationsConfig]{
	Default: &integrationsConfig{},
	Dev:     &integrationsConfig{Test: true},
}

func integration[T config.ComponentConfig](name string, cfg config.Set[T], init any) fx.Option {
	return Component(name, cfg, fx.Provide(
		fx.Annotate(
			init,
			fx.ResultTags(`group:"integrations"`),
		),
	))
}

func integrationsFXOption() fx.Option {
	return fx.Options(
		integration("asana", config.EmptySet, asana.New),
		integration("aws", config.EmptySet, aws.New),
		integration("calendar", config.EmptySet, calendar.New),
		integration("chatgpt", config.EmptySet, chatgpt.New),
		integration("confluence", config.EmptySet, confluence.New),
		integration("discord", config.EmptySet, discord.New),
		integration("drive", config.EmptySet, drive.New),
		integration("forms", config.EmptySet, forms.New),
		integration("github", config.EmptySet, github.New),
		integration("gmail", config.EmptySet, gmail.New),
		integration("gemini", config.EmptySet, gemini.New),
		integration("google", config.EmptySet, google.New),
		integration("grpc", config.EmptySet, grpc.New),
		integration("jira", config.EmptySet, jira.New),
		integration("redis", config.EmptySet, redis.New),
		integration("sheets", config.EmptySet, sheets.New),
		integration("slack", config.EmptySet, slack.New),
		integration("twilio", config.EmptySet, twilio.New),
		fx.Invoke(func(lc fx.Lifecycle, l *zap.Logger, muxes *muxes.Muxes, svcs sdkservices.Services) {
			HookOnStart(lc, func(ctx context.Context) error {
				asana.Start(l, muxes)
				aws.Start(l, muxes)
				chatgpt.Start(l, muxes)
				confluence.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
				discord.Start(l, muxes, svcs.Vars(), svcs.Dispatcher())
				github.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
				gemini.Start(l, muxes)
				google.Start(l, muxes, svcs.Vars(), svcs.OAuth(), svcs.Dispatcher())
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
