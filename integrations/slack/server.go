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

// Start initializes all the HTTP handlers of the Slack integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, noAuth *http.ServeMux, auth *http.ServeMux, v sdkservices.Vars, d sdkservices.Dispatcher) {
	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	noAuth.Handle(uiPath, http.FileServer(http.FS(static.SlackWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	auth.Handle("GET "+oauthPath, NewHandler(l))

	wsh := websockets.NewHandler(l, v, d, desc)
	auth.HandleFunc("POST "+formPath, wsh.HandleForm)

	// Event webhooks (unauthenticated by definition).
	whh := webhooks.NewHandler(l, v, d, integrationID)
	noAuth.HandleFunc("POST "+webhooks.BotEventPath, whh.HandleBotEvent)
	noAuth.HandleFunc("POST "+webhooks.SlashCommandPath, whh.HandleSlashCommand)
	noAuth.HandleFunc("POST "+webhooks.InteractionPath, whh.HandleInteraction)

	// Initialize WebSocket pool.
	cids, err := v.FindConnectionIDs(context.Background(), integrationID, vars.AppTokenName, "")
	if err != nil {
		l.Error("Failed to list WebSocket-based connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		data, err := v.Get(context.Background(), sdktypes.NewVarScopeID(cid))
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
