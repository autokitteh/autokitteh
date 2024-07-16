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

	// savePath is the URL path for our handler to save a new API token / PAT
	// connection, after the user submits its details via a web form.
	savePath = "/jira/save"

	// WebhookPath is the URL path for our webhook to handle asynchronous events.
	webhookPath = "/jira/webhook"
)

func Start(l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI + handlers.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.Handle(uiPath, http.FileServer(http.FS(static.JiraWebContent)))

	h := NewHTTPHandler(l, o, vars, d)
	mux.HandleFunc("GET "+oauthPath, h.handleOAuth)
	mux.HandleFunc("POST "+savePath, h.handleSave)

	// Event webhook.
	mux.HandleFunc("POST "+webhookPath, h.handleEvent)
}
