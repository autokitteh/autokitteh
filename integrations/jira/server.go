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
)

func Start(l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI + handlers.
	mux.Handle(desc.ConnectionURL().Path, http.FileServer(http.FS(static.JiraWebContent)))

	h := NewHTTPHandler(l, o)
	mux.HandleFunc(oauthPath, h.handleOAuth)

	// TODO(ENG-965): Event webhooks.
}
