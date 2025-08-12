package hubspot

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the HubSpot integration.
// This includes connection UIs and initialization webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, o *oauth.OAuth, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.HubSpotWebContent)

	h := NewHTTPHandler(l, v, o, d)

	common.RegisterOAuthHandler(m, desc, h.ServeHTTP)
	pattern := fmt.Sprintf("%s %s/webhook", http.MethodPost, desc.ConnectionURL().Path)
	m.NoAuth.HandleFunc(pattern, h.handleEvent)
}
