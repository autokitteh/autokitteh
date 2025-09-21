package notion

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(cvars),
		connTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// Start initializes all the HTTP handlers of the integration.
// This includes an internal connection UI, webhooks for AutoKitteh
// connection initialization, and asynchronous event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars) {
	h := NewHTTPHandler(l, v)
	common.RegisterSaveHandler(m, desc, h.handleSave)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)
}

// handler implements several HTTP webhooks to save authentication data.
type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
}

func NewHTTPHandler(l *zap.Logger, v sdkservices.Vars) handler {
	l = l.With(zap.String("integration", desc.UniqueName().String()))
	return handler{logger: l, vars: v}
}
