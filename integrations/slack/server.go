package slack

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/integrations/slack/websockets"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// formPath is the URL path for our handler to save a new
	// autokitteh connection, based on a user-submitted form.
	formPath = "/slack/save"

	// oauthPath is the URL path for our handler to save
	// new OAuth-based connections.
	oauthPath = "/slack/oauth"
)

func Start(l *zap.Logger, mux *http.ServeMux, vs sdkservices.Vars, d sdkservices.Dispatcher) {
	// Connection UI + save handlers.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.Handle(uiPath, http.FileServer(http.FS(static.SlackWebContent)))

	mux.Handle("GET "+oauthPath, NewHandler(l))

	wsh := websockets.NewHandler(l, vs, d, integrationID)
	mux.HandleFunc("POST "+formPath, wsh.HandleForm)

	// Event webhooks.
	whh := webhooks.NewHandler(l, vs, d, integrationID)
	mux.HandleFunc("POST "+webhooks.BotEventPath, whh.HandleBotEvent)
	mux.HandleFunc("POST "+webhooks.SlashCommandPath, whh.HandleSlashCommand)
	mux.HandleFunc("POST "+webhooks.InteractionPath, whh.HandleInteraction)

	// Initialize WebSocket pool.
	cids, err := vs.FindConnectionIDs(context.Background(), integrationID, vars.AppTokenName, "")
	if err != nil {
		l.Error("Failed to list WebSocket-based connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		data, err := vs.Reveal(context.Background(), sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("Missing data for Slack Socket Mode app", zap.Error(err))
			continue
		}

		var vs vars.Vars
		data.Decode(&vs)
		appToken := data.GetValue(vars.AppTokenName)
		botToken := data.GetValue(vars.BotTokenName)

		wsh.OpenSocketModeConnection(vs.AppID, botToken, appToken)
	}
}
