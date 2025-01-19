package twilio

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/twilio/webhooks"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the Twilio integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	h := webhooks.NewHandler(l, v, d, "twilio", desc)

	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	muxes.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.TwilioWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	muxes.Auth.HandleFunc(webhooks.AuthPath, h.HandleAuth)

	// Event webhook (unauthenticated by definition).
	muxes.NoAuth.HandleFunc(webhooks.MessagePath, h.HandleMessage)
}
