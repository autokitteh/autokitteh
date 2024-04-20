package svc

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/aws"
	"go.autokitteh.dev/autokitteh/integrations/chatgpt"
	"go.autokitteh.dev/autokitteh/integrations/github"
	"go.autokitteh.dev/autokitteh/integrations/google"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/sheets"
	"go.autokitteh.dev/autokitteh/integrations/grpc"
	httpint "go.autokitteh.dev/autokitteh/integrations/http"
	"go.autokitteh.dev/autokitteh/integrations/redis"
	"go.autokitteh.dev/autokitteh/integrations/scheduler"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func intg[T any](n string, cfg configset.Set[T], f any, iopts ...fx.Option) fx.Option {
	iopts = append([]fx.Option{fx.Provide(
		fx.Annotate(
			f,
			fx.As(new(sdkservices.Integration)),
			fx.ResultTags(`group:"integrations"`),
		),
	)}, iopts...)

	return Component(n, cfg, iopts...)
}

func newIntegrationsOpts(ropts RunOptions) (opts []fx.Option) {
	if ropts.Mode.IsDev() || ropts.Mode.IsTest() {
		opts = append(opts, intg("test", configset.Empty, newTestIntegration))
	}

	opts = append(
		opts,

		intg("aws", configset.Empty, aws.New),
		intg("github", configset.Empty, github.New, fx.Invoke(github.InitServer)),
		intg("chatgpt", configset.Empty, chatgpt.New, fx.Invoke(chatgpt.InitServer)),
		intg("gmail", configset.Empty, gmail.New),
		intg("google", configset.Empty, google.New, fx.Invoke(google.InitServer)),
		intg("sheets", configset.Empty, sheets.New),
		intg("http", configset.Empty, httpint.New, fx.Invoke(httpint.InitServer)),
		intg("redis", configset.Empty, redis.New),
		intg("twilio", configset.Empty, twilio.New, fx.Invoke(twilio.InitServer)),
		intg("grpc", configset.Empty, grpc.New),

		intg(
			"slack",
			configset.Empty,
			slack.New,
			fx.Invoke(
				slack.InitServer,
				fx.Annotate(slack.InitEventSource, fx.OnStart(
					func(s *slack.EventSource) { s.Start() },
				)),
			),
		),

		intg(
			"scheduler",
			configset.Empty,
			scheduler.New,
			fx.Invoke(scheduler.InitServer), fx.Invoke(
				func(lc fx.Lifecycle, l *zap.Logger, s sdkservices.Secrets, d sdkservices.Dispatcher) {
					HookOnStart(lc, func(ctx context.Context) error {
						scheduler.StartEventSource(l, s, d)
						return nil
					})
				},
			),
		),

		fx.Provide(fx.Annotate(sdkintegrations.New, fx.ParamTags(`group:"integrations"`))),
	)

	return
}
