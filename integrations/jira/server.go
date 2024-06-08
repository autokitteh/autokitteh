package jira

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users at
	// the beginning and the end of their 3-legged OAuth 2.0 flow with Atlassian.
	uiPath = "/jira/connect/"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/jira/oauth"
)

func Start(l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI + handler.
	mux.Handle(uiPath, http.FileServer(http.FS(static.JiraWebContent)))

	// TODO(ENG-965): Implement OAuth handler.
}
