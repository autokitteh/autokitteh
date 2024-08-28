package twilio

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/twilio/webhooks"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the Twilio integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, noAuth *http.ServeMux, auth *http.ServeMux, v sdkservices.Vars, d sdkservices.Dispatcher) {
	h := webhooks.NewHandler(l, v, d, "twilio", desc)

	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	noAuth.Handle(uiPath, http.FileServer(http.FS(static.TwilioWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	auth.HandleFunc(webhooks.AuthPath, h.HandleAuth)

	// Event webhook (unauthenticated by definition).
	noAuth.HandleFunc(webhooks.MessagePath, h.HandleMessage)
}
