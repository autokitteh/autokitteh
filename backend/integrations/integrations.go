package integrations

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/aws"
	"go.autokitteh.dev/autokitteh/integrations/chatgpt"
	"go.autokitteh.dev/autokitteh/integrations/github"
	"go.autokitteh.dev/autokitteh/integrations/google"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/sheets"
	httpint "go.autokitteh.dev/autokitteh/integrations/http"
	"go.autokitteh.dev/autokitteh/integrations/redis"
	"go.autokitteh.dev/autokitteh/integrations/scheduler"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func New(s sdkservices.Secrets) sdkservices.Integrations {
	return sdkintegrations.New([]sdkservices.Integration{
		aws.New(), // TODO: Secrets
		chatgpt.New(s),
		github.New(s),
		gmail.New(s),
		google.New(s),
		sheets.New(s),
		// TODO: gRPC
		httpint.New(s),
		redis.New(), // TODO: Secrets
		scheduler.New(s),
		slack.New(s),
		twilio.New(s),
	})
}

func Start(_ context.Context, l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d sdkservices.Dispatcher) error {
	// TODO: AWS
	chatgpt.Start(l, mux, s)
	github.Start(l, mux, s, o, d)
	google.Start(l, mux, s, o, d)
	httpint.Start(l, mux, d)
	// TODO: ProxySQL
	// TODO: Redis
	scheduler.Start(l, mux, s, d)
	slack.Start(l, mux, s, d)
	twilio.Start(l, mux, s, d)

	return nil
}
