package kubernetes

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
)

const (
	// savePath is the URL path for our handler to save new
	// connections, after users submit them via a web form.
	savePath = "/kubernetes/save"
)

func Start(l *zap.Logger, m *muxes.Muxes) {
	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	m.Auth.Handle("POST "+savePath, NewHTTPHandler(l))
}
