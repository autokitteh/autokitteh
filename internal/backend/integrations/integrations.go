package integrations

import (
	"context"
	"net/http"

	"go.uber.org/zap"

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
	httpint "go.autokitteh.dev/autokitteh/integrations/http"
	"go.autokitteh.dev/autokitteh/integrations/redis"
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

func New(cfg *Config, vars sdkservices.Vars) sdkservices.Integrations {
	ints := []sdkservices.Integration{
		aws.New(vars),
		calendar.New(vars),
		chatgpt.New(vars),
		confluence.New(vars),
		discord.New(vars),
		drive.New(vars),
		forms.New(vars),
		github.New(vars),
		gmail.New(vars),
		gemini.New(vars),
		google.New(vars),
		grpc.New(),
		httpint.New(vars),
		jira.New(vars),
		redis.New(vars),
		sheets.New(vars),
		slack.New(vars),
		twilio.New(vars),
	}

	if cfg.Test {
		ints = append(ints, newTestIntegration())
	}

	return sdkintegrations.New(ints)
}

func Start(_ context.Context, l *zap.Logger, muxNoAuth *http.ServeMux, muxAuth *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher, c sdkservices.Connections, p sdkservices.Projects) error {
	aws.Start(l, muxNoAuth)
	chatgpt.Start(l, muxNoAuth)
	confluence.Start(l, muxNoAuth, vars, o, d)
	discord.Start(l, muxNoAuth)
	github.Start(l, muxNoAuth, vars, o, d)
	gemini.Start(l, muxNoAuth)
	google.Start(l, muxNoAuth, muxAuth, vars, o, d)
	httpint.Start(l, muxNoAuth, d, c, p)
	jira.Start(l, muxNoAuth, vars, o, d)
	slack.Start(l, muxNoAuth, vars, d)
	twilio.Start(l, muxNoAuth, vars, d)

	return nil
}
