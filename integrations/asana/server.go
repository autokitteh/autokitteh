package asana

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users.
	uiPath = "/asana/connect/"

	// savePath is the URL path for our handler to save new
	// connections, after users submit them via a web form.
	savePath = "/asana/save"
)

// Start initializes all the HTTP handlers of the AWS integration.
// This includes connection UIs and initialization webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes) {
	// Connection UI.
	muxes.Main.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.AsanaWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	muxes.Main.Auth.Handle(savePath, NewHTTPHandler(l))
}
