package microsoft

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of all the Microsoft integrations. This
// includes connection UIs, connection initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) {
	// Connection UI for authenticated AutoKitteh users (user authentication
	// isn't required, but it makes no sense to create a connection without it).
	muxes.Auth.Handle("GET /microsoft/", http.FileServer(http.FS(static.MicrosoftWebContent)))
}
