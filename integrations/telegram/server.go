package telegram

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/telegram/webhooks"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the Telegram integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	// Serve static UI for the connection page
	common.ServeStaticUI(m, desc, static.TelegramWebContent)

	// Register save handler for connection initialization
	saveHandler := NewHTTPHandler(l, v)
	common.RegisterSaveHandler(m, desc, saveHandler.ServeHTTP)

	// Register webhook handler for Telegram updates (event webhooks, no auth)
	iid := sdktypes.NewIntegrationIDFromName(desc.UniqueName().String())
	whHandler := webhooks.NewHandler(l, v, d, iid)
	m.NoAuth.HandleFunc("POST "+webhooks.UpdatePath, whHandler.HandleUpdate)
}
