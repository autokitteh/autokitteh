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
	formPath = "/slack/save_tokens"
)

func Start(l *zap.Logger, mux *http.ServeMux, vs sdkservices.Vars, d sdkservices.Dispatcher) {
	// Connection UI + save handlers.
	wsh := websockets.NewHandler(l, vs, d, integrationID)
	mux.Handle(uiPath, http.FileServer(http.FS(static.SlackWebContent)))
	mux.Handle(oauthPath, NewHandler(l, vs))
	mux.HandleFunc(formPath, wsh.HandleForm)

	// Event webhooks.
	whh := webhooks.NewHandler(l, vs, d, integrationID)
	mux.HandleFunc(webhooks.BotEventPath, whh.HandleBotEvent)
	mux.HandleFunc(webhooks.SlashCommandPath, whh.HandleSlashCommand)
	mux.HandleFunc(webhooks.InteractionPath, whh.HandleInteraction)

	// Initialize WebSocket pool.
	cids, err := vs.FindConnectionIDs(context.Background(), integrationID, vars.WebSocketName, "")
	if err != nil {
		l.Error("Failed to list WebSocket cids", zap.Error(err))
		return
	}

	for _, cid := range cids {
		data, err := vs.Get(context.Background(), sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("Missing data for Slack Socket Mode app", zap.Error(err))
			continue
		}

		var vs vars.Vars
		data.Decode(&vs)

		wsh.OpenSocketModeConnection(vs.AppID, data.GetValue(vars.BotTokenName), vs.AppToken)
	}
}
