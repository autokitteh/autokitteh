package twilio

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/twilio/webhooks"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the Twilio integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.TwilioWebContent)

	h := webhooks.NewHandler(l, v, d, "twilio", desc)
	common.RegisterSaveHandler(m, desc, h.HandleAuth)

	// Event webhook (unauthenticated by definition).
	m.NoAuth.HandleFunc(webhooks.MessagePath, h.HandleMessage)
}
