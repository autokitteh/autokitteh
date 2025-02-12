package linear

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the Linear integration. This
// includes connection UIs, connection initialization webhooks, and event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.LinearWebContent)

	h := newHTTPHandler(l, v, o, d)
	common.RegisterSaveHandler(m, desc, h.handleSave)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)

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
