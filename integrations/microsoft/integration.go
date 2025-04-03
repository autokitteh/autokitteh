package microsoft

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/connection"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/teams"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

var desc = common.Descriptor("microsoft", "Microsoft (All APIs)", "/static/images/microsoft.svg")

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(v sdkservices.Vars, o *oauth.OAuth) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc, sdkmodule.New(),
		connection.Status(v), connection.Test(v, o),
		sdkintegrations.WithConnectionConfigFromVars(v))
}

// Start initializes all the HTTP handlers of the integration.
// This includes an internal connection UI, webhooks for AutoKitteh
// connection initialization, and asynchronous event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, o *oauth.OAuth, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.MicrosoftWebContent)

	h := newHTTPHandler(l, v, o, d)
	common.RegisterSaveHandler(m, desc, h.handleSave)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)

	// Event webhooks (no AutoKitteh user authentication by definition, because
	// these asynchronous requests are sent to us by third-party services).
	m.NoAuth.HandleFunc("POST "+connection.ChangePath, h.handleEvent)
	m.NoAuth.HandleFunc("POST "+connection.LifecyclePath, h.handleLifecycle)
}

// handler implements several HTTP webhooks to save authentication data, as
// well as receive and dispatch third-party asynchronous event notifications.
type handler struct {
	logger   *zap.Logger
	vars     sdkservices.Vars
	oauth    *oauth.OAuth
	dispatch sdkservices.DispatchFunc
}

func newHTTPHandler(l *zap.Logger, v sdkservices.Vars, o *oauth.OAuth, d sdkservices.DispatchFunc) handler {
	l = l.With(zap.String("integration", desc.UniqueName().String()))
	return handler{logger: l, oauth: o, vars: v, dispatch: d}
}

// resources returns the Microsoft Graph resources that each Microsoft integration
// should subscribe to in order to receive asynchronous change notifications.
func resources(i sdktypes.Integration) []string {
	// TODO: Convert this to a switch when we add more integrations.
	if i.UniqueName().String() == teams.IntegrationName {
		return teams.SubscriptionResources
	}
	return nil
}
