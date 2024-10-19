package github

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/webhooks"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/github/connect"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/github/oauth"

	// savePath is the URL path for our handler to save new PAT-based
	// connections, after users submit them via a web form.
	savePath = "/github/save"
)

// Start initializes all the HTTP handlers of the GitHub integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	muxes.Main.NoAuth.HandleFunc(uiPath, connect.ServeHTTP)
	muxes.Main.NoAuth.Handle(uiPath+"{filename}", http.FileServer(http.FS(static.GitHubWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	h := NewHandler(l, o)
	muxes.Main.Auth.HandleFunc("GET "+oauthPath, h.handleOAuth)
	muxes.Main.Auth.HandleFunc("POST "+savePath, h.handlePAT)

	// Event webhooks (unauthenticated by definition).
	eventHandler := webhooks.NewHandler(l, v, d, integrationID)
	muxes.Main.NoAuth.Handle("POST "+webhooks.WebhookPath+"/{id}", eventHandler) // User events.
	muxes.Main.NoAuth.Handle("POST "+webhooks.WebhookPath, eventHandler)         // App events.
}
