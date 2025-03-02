package aws

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/web/static"
)

// Start initializes all the HTTP handlers of the AWS integration.
// This includes connection UIs and initialization webhooks.
func Start(l *zap.Logger, m *muxes.Muxes) {
	common.ServeStaticUI(m, desc, static.AWSWebContent)

	common.RegisterSaveHandler(m, desc, NewHTTPHandler(l).ServeHTTP)
}
