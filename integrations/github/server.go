package github

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/webhooks"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/github/connect"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, mux *http.ServeMux, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI + handlers.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.HandleFunc(uiPath, connect.ServeHTTP)
	mux.Handle(uiPath+"{filename}", http.FileServer(http.FS(static.GitHubWebContent)))

	h := NewHandler(l, o)
	mux.HandleFunc("GET "+oauthPath, h.handleOAuth)
	mux.HandleFunc("POST "+patPath, h.handlePAT)

	// Event webhooks.
	// TODO: Use Go 1.22's pattern wildcards to have 2 separate event
	// handler functions (https://go.dev/blog/routing-enhancements).
	eventHandler := webhooks.NewHandler(l, v, d, integrationID)
	mux.Handle("POST "+webhooks.WebhookPath+"/", eventHandler) // User events.
	mux.Handle("POST "+webhooks.WebhookPath, eventHandler)     // App events.
}
