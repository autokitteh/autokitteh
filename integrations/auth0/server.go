package auth0

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// oauthPath is the URL path for our handler to save
	// new OAuth-based connections.
	oauthPath = "/auth0/oauth"

	// savePath is the URL path for our handler to save authentication credentials.
	// It acts as a passthrough for the OAuth flow.
	savePath = "/auth0/save"
)

// Start initializes all the HTTP handlers of the Auth0 integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, vars sdkservices.Vars) {
	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	muxes.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.Auth0WebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	muxes.Auth.Handle("GET "+oauthPath, NewHandler(l, vars))
	muxes.Auth.Handle("GET "+savePath, NewHandler(l, vars))
}
