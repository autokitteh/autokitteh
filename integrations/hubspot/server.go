package hubspot

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the HubSpot integration.
// This includes connection UIs and initialization webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, o *oauth.OAuth) {
	common.ServeStaticUI(m, desc, static.HubSpotWebContent)

	common.RegisterOAuthHandler(m, desc, NewHTTPHandler(l, o).ServeHTTP)
}
