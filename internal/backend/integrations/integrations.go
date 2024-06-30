package integrations

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/atlassian/confluence"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/jira"
	"go.autokitteh.dev/autokitteh/integrations/aws"
	"go.autokitteh.dev/autokitteh/integrations/chatgpt"
	"go.autokitteh.dev/autokitteh/integrations/github"
	"go.autokitteh.dev/autokitteh/integrations/google"
	"go.autokitteh.dev/autokitteh/integrations/google/calendar"
	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/forms"
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
		drive.New(vars),
		forms.New(vars),
		github.New(vars),
		gmail.New(vars),
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

func Start(_ context.Context, l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher, c sdkservices.Connections, p sdkservices.Projects) error {
	aws.Start(mux)
	chatgpt.Start(l, mux)
	confluence.Start(l, mux, vars, o, d)
	github.Start(l, mux, vars, o, d)
	google.Start(l, mux, o, d)
	httpint.Start(l, mux, d, c, p)
	jira.Start(l, mux, vars, o, d)
	slack.Start(l, mux, vars, d)
	twilio.Start(l, mux, vars, d)

	return nil
}
