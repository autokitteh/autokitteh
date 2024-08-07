package twilio

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/twilio/webhooks"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, d sdkservices.Dispatcher) {
	h := webhooks.NewHandler(l, vars, d, "twilio", desc)

	// Save new autokitteh connections with user-submitted Twilio secrets.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.Handle(uiPath, http.FileServer(http.FS(static.TwilioWebContent)))

	mux.HandleFunc(webhooks.AuthPath, h.HandleAuth)

	// Event webhooks.
	mux.HandleFunc(webhooks.MessagePath, h.HandleMessage)
}
