package telegram

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(i),
		connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// Start initializes all the HTTP handlers of the Telegram integration.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	h := NewHandler(l, v, d)
	common.RegisterSaveHandler(m, desc, h.handleSave)

	// Webhook handler for receiving Telegram events.
	pattern := fmt.Sprintf("%s %s/webhook/{connection_id}", http.MethodPost, desc.ConnectionURL().Path)
	m.NoAuth.HandleFunc(pattern, h.handleEvent)
}

type handler struct {
	logger   *zap.Logger
	vars     sdkservices.Vars
	dispatch sdkservices.DispatchFunc
}

// NewHandler creates a new webhook handler for Telegram events.
func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.DispatchFunc) handler {
	return handler{
		logger:   l.With(zap.String("integration", desc.UniqueName().String())),
		vars:     v,
		dispatch: d,
	}
}
