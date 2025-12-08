package discord

import (
	"context"
	"fmt"
	"os"
	"strconv"
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

	// pollInterval is the interval at which we poll the database
	// for active connections and ensure WebSockets are open.
	pollInterval = 5 * time.Minute
)

// Start initializes all the HTTP handlers of the Discord integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.DiscordWebContent)

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	wsh := NewHandler(l, v, d, desc)

	m.Auth.Handle("POST "+savePath, wsh)

	// Initialize WebSocket pool.
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

	// Start background polling to ensure WebSocket connections exist.
	go pollAndSyncWebSockets(context.Background(), l, v, wsh)
}

// pollAndSyncWebSockets runs a periodic check to ensure all active connections
// have WebSocket connections open.
func pollAndSyncWebSockets(ctx context.Context, l *zap.Logger, v sdkservices.Vars, wsh handler) {
	interval := getPollingInterval(l)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.Info("Stopping Discord WebSocket polling due to context cancellation")
			return
		case <-ticker.C:
			syncWebSocketConnections(ctx, l, v, wsh)
		}
	}
}

// syncWebSocketConnections queries the database for all active connections
// and ensures WebSocket connections are open for them.
func syncWebSocketConnections(ctx context.Context, l *zap.Logger, v sdkservices.Vars, wsh handler) {
	cids, err := v.FindActiveConnectionIDs(ctx, integrationID, vars.BotToken, "")
	if err != nil {
		l.Error("Failed to list active Discord connection IDs during sync "+err.Error(), zap.Error(err))
		return
	}

	l.Debug("Syncing Discord WebSocket connections", zap.Int("connection_count", len(cids)))

	for _, cid := range cids {
		data, err := v.Get(ctx, sdktypes.NewVarScopeID(cid))
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

func getPollingInterval(l *zap.Logger) time.Duration {
	interval := pollInterval

	if envInterval := os.Getenv("DISCORD_POLL_INTERVAL_MINUTES"); envInterval != "" {
		if minutes, err := strconv.Atoi(envInterval); err == nil && minutes > 0 {
			interval = time.Duration(minutes) * time.Minute
			l.Info("Using custom Discord WebSocket poll interval", zap.Duration("interval", interval))
		} else {
			l.Warn("Invalid DISCORD_POLL_INTERVAL_MINUTES, using default", zap.String("value", envInterval))
		}
	}

	return interval
}
