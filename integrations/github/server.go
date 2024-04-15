package github

import (
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/github/connect"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Connection UI + handler.
	mux.HandleFunc(uiPath, connect.ServeHTTP)
	staticFiles := http.FileServer(http.FS(static.GitHubWebContent))
	mux.Handle(kittehs.Must1(url.JoinPath(uiPath, "error.html")), staticFiles)
	mux.Handle(kittehs.Must1(url.JoinPath(uiPath, "error.js")), staticFiles)
	mux.Handle(kittehs.Must1(url.JoinPath(uiPath, "form.js")), staticFiles)
	mux.Handle(kittehs.Must1(url.JoinPath(uiPath, "styles.css")), staticFiles)
	mux.Handle(kittehs.Must1(url.JoinPath(uiPath, "success.html")), staticFiles)
	mux.Handle(kittehs.Must1(url.JoinPath(uiPath, "success.js")), staticFiles)
	h := NewHandler(l, s, o, "github")
	mux.HandleFunc(patPath, h.handlePAT)
	mux.HandleFunc(oauthPath, h.handleOAuth)

	// Event webhooks.
	eventHandler := webhooks.NewHandler(l, s, d, "github", integrationID)
	mux.Handle(webhooks.WebhookPath+"/", eventHandler) // User events.
	mux.Handle(webhooks.WebhookPath, eventHandler)     // App events.
}
