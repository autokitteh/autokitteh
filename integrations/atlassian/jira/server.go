package jira

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/jira/oauth"

	// savePath is the URL path for our handler to save new token-based
	// connections, after users submit them via a web form.
	savePath = "/jira/save"

	// WebhookPath is the URL path for our webhook to handle asynchronous events.
	webhookPath = "/jira/webhook"
)

// Start initializes all the HTTP handlers of the Jira integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) {
	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	muxes.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.JiraWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	h := NewHTTPHandler(l, o, v, d)
	muxes.Auth.HandleFunc("GET "+oauthPath, h.handleOAuth)
	muxes.Auth.HandleFunc("POST "+savePath, h.handleSave)

	// Event webhook (unauthenticated by definition).
	muxes.NoAuth.HandleFunc("POST "+webhookPath, h.handleEvent)
}
