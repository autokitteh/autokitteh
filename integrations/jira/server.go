package jira

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/jira/oauth"

	// WebhookPath is the URL path for our webhook to handle asynchronous events.
	webhookPath = "/jira/webhook"
)

func Start(l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI + handlers.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.Handle(uiPath, http.FileServer(http.FS(static.JiraWebContent)))

	h := NewHTTPHandler(l, o, vars, d)
	mux.HandleFunc("GET "+oauthPath, h.handleOAuth)

	// Event webhook.
	mux.HandleFunc("POST "+webhookPath, h.handleEvent)
}
