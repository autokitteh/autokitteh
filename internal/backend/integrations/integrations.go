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

type Config struct {
	Test bool `koanf:"test"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{Test: true},
}

func New(cfg *Config, s sdkservices.Secrets) sdkservices.Integrations {
	ints := []sdkservices.Integration{
		aws.New(), // TODO: Secrets
		chatgpt.New(s),
		github.New(s),
		gmail.New(s),
		google.New(s),
		// TODO: gRPC
		httpint.New(s),
		redis.New(), // TODO: Secrets
		scheduler.New(s),
		sheets.New(s),
		slack.New(s),
		twilio.New(s),
		grpc.New(s),
	}

	if cfg.Test {
		ints = append(ints, newTestIntegration())
	}

	return sdkintegrations.New(ints)
}

func Start(_ context.Context, l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d sdkservices.Dispatcher) error {
	// TODO: AWS
	chatgpt.Start(l, mux, s)
	github.Start(l, mux, s, o, d)
	google.Start(l, mux, s, o, d)
	httpint.Start(l, mux, s, d)
	// TODO: Redis
	scheduler.Start(l, mux, s, d)
	slack.Start(l, mux, s, d)
	twilio.Start(l, mux, s, d)

	return nil
}
