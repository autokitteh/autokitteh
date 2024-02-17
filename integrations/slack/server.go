package slack

import (
	"net/http"
	"os"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, d sdkservices.Dispatcher) {
	if !checkRequiredEnvVars(l) {
		return
	}

	// Connection UI + handler.
	mux.Handle(uiPath, http.FileServer(http.FS(static.SlackWebContent)))
	mux.Handle(oauthPath, NewHandler(l, s, "slack"))

	// Event webhooks.
	h := webhooks.NewHandler(l, s, d, "slack", integrationID)
	mux.HandleFunc(webhooks.BotEventPath, h.HandleBotEvent)
	mux.HandleFunc(webhooks.SlashCommandPath, h.HandleSlashCommand)
	mux.HandleFunc(webhooks.InteractionPath, h.HandleInteraction)
}

func checkRequiredEnvVars(l *zap.Logger) bool {
	result := true
	for _, k := range []string{
		// OAuth
		"SLACK_APP_ID",
		"SLACK_CLIENT_ID",
		"SLACK_CLIENT_SECRET",
		// webhooks/webhook.go
		"SLACK_SIGNING_SECRET",
	} {
		if os.Getenv(k) == "" {
			l.Warn("Required environment variable is missing",
				zap.String("name", k),
			)
			result = false
		}
	}
	return result
}
