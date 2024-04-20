package slack

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/integrations/slack/websockets"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// formPath is the URL path for our handler to save a new
	// autokitteh connection, based on a user-submitted form.
	formPath = "/slack/save_tokens"
)

func InitServer(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, d sdkservices.Dispatcher) {
	// Event webhooks.
	whh := webhooks.NewHandler(l, s, d, "slack", integrationID)
	mux.HandleFunc(webhooks.BotEventPath, whh.HandleBotEvent)
	mux.HandleFunc(webhooks.SlashCommandPath, whh.HandleSlashCommand)
	mux.HandleFunc(webhooks.InteractionPath, whh.HandleInteraction)

	// Connection UI + save handlers.
	mux.Handle(uiPath, http.FileServer(http.FS(static.SlackWebContent)))
	mux.Handle(oauthPath, NewHandler(l, s, "slack"))
}

type EventSource struct {
	l                        *zap.Logger
	mux                      *http.ServeMux
	s                        sdkservices.Secrets
	d                        sdkservices.Dispatcher
	openSocketModeConnection func(appID, botToken, appToken string)
}

func InitEventSource(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, d sdkservices.Dispatcher) *EventSource {
	wsh := websockets.NewHandler(l, s, d, "slack", integrationID)
	mux.HandleFunc(formPath, wsh.HandleForm)

	return &EventSource{
		l:                        l,
		mux:                      mux,
		s:                        s,
		d:                        d,
		openSocketModeConnection: wsh.OpenSocketModeConnection,
	}
}

func (s *EventSource) Start() {
	// Initialize WebSocket pool.
	tokens, err := s.s.List(context.Background(), "slack", "websockets")
	if err != nil {
		s.l.Error("Failed to list WebSocket tokens", zap.Error(err))
		return
	}

	for _, connToken := range tokens {
		data, err := s.s.Get(context.Background(), "slack", connToken)
		if err != nil {
			s.l.Error("Missing data for Slack Socket Mode app", zap.Error(err))
			continue
		}

		s.openSocketModeConnection(data["appID"], data["botToken"], data["appLevelToken"])
	}
}
