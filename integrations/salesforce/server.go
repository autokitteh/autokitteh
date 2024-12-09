package salesforce

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/web/static"
)

// oauthPath is the URL path for our handler to save
// new OAuth-based connections.
const oauthPath = "/salesforce/oauth"

// Start initializes all the HTTP handlers of the Salesforce integration.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars) {
	// Connection UI
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	muxes.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.SalesforceWebContent)))

	muxes.Auth.Handle("GET "+oauthPath, NewHandler(l))
}
