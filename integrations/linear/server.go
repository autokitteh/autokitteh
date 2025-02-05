package linear

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the Linear integrations. This
// includes connection UIs, connection initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) {
	// Connection UI for authenticated AutoKitteh users (user authentication
	// isn't required, but it makes no sense to create a connection without it).
	muxes.Auth.Handle("GET /linear/", http.FileServer(http.FS(static.LinearWebContent)))

	// // Connection initialization webhooks save connection variables (e.g. auth and
	// // metadata), which requires an authenticated user context for database access.
	h := newHTTPHandler(l, v, o, d)
	muxes.Auth.HandleFunc("POST /linear/save", h.handleSave)
	muxes.Auth.HandleFunc("GET /linear/save", h.handleSave)
	muxes.Auth.HandleFunc("GET /linear/oauth", h.handleOAuth)

	// TODO: Event webhooks (no AutoKitteh user authentication by definition, because
	// these asynchronous requests are sent to us by third-party services).
}

// handler implements several HTTP webhooks to save authentication data, as
// well as receive and dispatch third-party asynchronous event notifications.
type handler struct {
	logger   *zap.Logger
	vars     sdkservices.Vars
	oauth    sdkservices.OAuth
	dispatch sdkservices.DispatchFunc
}

func newHTTPHandler(l *zap.Logger, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) handler {
	return handler{logger: l, oauth: o, vars: v, dispatch: d}
}
