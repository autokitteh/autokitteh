package discord

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// savePath is the URL path for our handler to save new
	// connections, after users submit them via a web form.
	savePath = "/discord/save"
)

// Start initializes all the HTTP handlers of the Discord integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, d sdkservices.Dispatcher) {
	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	muxes.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.DiscordWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	wsh := NewHandler(l, v, d, desc)

	muxes.Auth.Handle("POST "+savePath, wsh)

	// Initialize WebSocket pool.
	cids, err := v.FindConnectionIDs(context.Background(), integrationID, vars.BotToken, "")
	if err != nil {
		l.Error("Failed to list WebSocket-based connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		data, err := v.Get(context.Background(), sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("Missing data for Discord WebSocket connection", zap.Error(err))
			continue
		}

		wsh.OpenWebSocketConnection(data.GetValue(vars.BotToken))
	}
}
