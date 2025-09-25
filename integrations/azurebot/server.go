package azurebot

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/azurebot/webhooks"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.AzureBotWebContent)

	h := webhooks.NewHandler(l, v, d, desc)
	common.RegisterSaveHandler(m, desc, h.HandleAuth)

	m.NoAuth.HandleFunc(webhooks.MessagePath, h.HandleMessage)
	m.NoAuth.HandleFunc(webhooks.ManifestPath, h.HandleManifest)
}
