package discord

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
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

// Config holds Discord-specific configuration.
type Config struct {
	EnableWebsockets bool `koanf:"enable_websockets"`
	PollInterval     int  `koanf:"poll_interval"` // in seconds
}

// StartWithConfig initializes all the HTTP handlers of the Discord integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func StartWithConfig(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc, cfg any) {
	// Extract config values
	enableWebsockets := false
	pollInterval := time.Duration(0)
	// Type assert to extract Discord config (comes from integrations registry)
	if discordCfg, ok := cfg.(Config); ok {
		enableWebsockets = discordCfg.EnableWebsockets
		pollInterval = time.Duration(discordCfg.PollInterval) * time.Second
	}

	common.ServeStaticUI(m, desc, static.DiscordWebContent)

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	wsh := NewHandler(l, v, d, desc)

	m.Auth.Handle("POST "+savePath, wsh)

	// Initialize WebSocket pool if enabled.
	if !enableWebsockets {
		l.Info("Discord WebSockets are disabled")
		return
	}

	cids, err := v.FindActiveConnectionIDs(context.Background(), integrationID, vars.BotToken, "")
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

	go pollAndSyncWebSockets(l, v, wsh, pollInterval)
}

// pollAndSyncWebSockets runs a periodic check to ensure all active connections
// have WebSocket connections open.
func pollAndSyncWebSockets(l *zap.Logger, v sdkservices.Vars, wsh handler, interval time.Duration) {
	l.Info("Starting Discord WebSocket polling", zap.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		syncWebSocketConnections(l, v, wsh)
	}
}

// syncWebSocketConnections queries the database for all active connections
// and ensures WebSocket connections are open for them.
func syncWebSocketConnections(l *zap.Logger, v sdkservices.Vars, wsh handler) {
	cids, err := v.FindActiveConnectionIDs(context.Background(), integrationID, vars.BotToken, "")
	if err != nil {
		l.Error("Failed to list active Discord connection IDs during sync "+err.Error(), zap.Error(err))
		return
	}

	l.Debug("Syncing Discord WebSocket connections", zap.Int("connection_count", len(cids)))

	for _, cid := range cids {
		data, err := v.Get(context.Background(), sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("Failed to get connection data during WebSocket sync "+err.Error(),
				zap.String("connection_id", cid.String()),
				zap.Error(err))
			continue
		}

		botToken := data.GetValue(vars.BotToken)
		if botToken == "" {
			l.Debug(fmt.Sprintf("Connection %s has empty bot token, skipping", cid.String()))
			continue
		}

		// OpenWebSocketConnection is idempotent - it checks the discordSessions
		// map and won't create a duplicate if one already exists for this bot token.
		wsh.OpenWebSocketConnection(botToken)
	}
}
