package github

import (
	"net/http"
	"net/url"
	"os"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/github/connect"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	if !checkRequiredEnvVars(l) {
		return
	}

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

func checkRequiredEnvVars(l *zap.Logger) bool {
	result := true
	for _, k := range []string{
		// OAuth
		"GITHUB_APP_NAME",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
		// oauth/jwt.go
		"GITHUB_PRIVATE_KEY",
		// webhooks/webhook.go
		"GITHUB_WEBHOOK_SECRET",
	} {
		if os.Getenv(k) == "" {
			l.Warn("Required environment variable is missing",
				zap.String("name", k),
			)
			result = false
		}
	}
	return result
}
