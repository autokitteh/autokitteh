package discord

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	savePath = "/discord/save"
)

// Start initializes all the HTTP handlers of the Discord integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, noAuth *http.ServeMux, auth *http.ServeMux) {
	// Connection UI.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	noAuth.Handle(uiPath, http.FileServer(http.FS(static.DiscordWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	auth.Handle("POST "+savePath, NewHTTPHandler(l))

	// TODO(ENG-1226): Event webhooks (unauthenticated by definition).
}
