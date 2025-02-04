package zoom

import (
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.uber.org/zap"
)

func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth) {
	// Connection UI for authenticated users
	// muxes.Auth.Handle("GET /zoom/", http.FileServer(http.FS(static.ZoomWebContent)))

	// Handler for Zoom OAuth integration.
	h := newHTTPHandler(l, v, o)

	// OAuth flow handlers.
	muxes.Auth.HandleFunc("POST /zoom/save", h.handleSave)  // Handles saving OAuth details.
	muxes.Auth.HandleFunc("GET /zoom/save", h.handleSave)   // Handles UI passthrough for OAuth.
	muxes.Auth.HandleFunc("GET /zoom/oauth", h.handleOAuth) // Handles Zoom OAuth redirection and token exchange.
}

type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
	oauth  sdkservices.OAuth
}

func newHTTPHandler(l *zap.Logger, v sdkservices.Vars, o sdkservices.OAuth) handler {
	return handler{logger: l, oauth: o, vars: v}
}
