package hubspot

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// oauthPath is the URL path for our handler to save new OAuth-based
	// connections, after users submit them via a web form.
	oauthPath = "/hubspot/oauth"
)

// Start initializes all the HTTP handlers of the HubSpot integration.
// This includes connection UIs and initialization webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, o sdkservices.OAuth) {
	common.ServeStaticUI(m, desc, static.HubSpotWebContent)

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	m.Auth.Handle(oauthPath, NewHTTPHandler(l, o))
}
