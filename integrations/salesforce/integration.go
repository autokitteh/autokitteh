package salesforce

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

var desc = common.Descriptor("salesforce", "Salesforce", "/static/images/salesforce.png")

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(v sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc, sdkmodule.New(), status(v), test(v),
		sdkintegrations.WithConnectionConfigFromVars(v))
}

// Start initializes all the HTTP handlers of the integration.
// This includes an internal connection UI, webhooks for AutoKitteh
// connection initialization, and asynchronous event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.SalesforceWebContent)

	h := newHTTPHandler(l, v, o, d)
	common.RegisterSaveHandler(m, desc, h.handleSave)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)

	// TODO: Event webhooks (no AutoKitteh user authentication by definition, because
	// these asynchronous requests are sent to us by third-party services).

	reopenExistingPubSubConnections(context.Background(), l, v, h)
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

func reopenExistingPubSubConnections(ctx context.Context, l *zap.Logger, v sdkservices.Vars, h handler) {
	iid := sdktypes.NewIntegrationIDFromName(desc.UniqueName().String())
	cids, err := v.FindConnectionIDs(ctx, iid, instanceURLVar, "")
	if err != nil {
		l.Error("failed to list Salesforce connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		data, err := v.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("can't restart Salesforce PubSub connection", zap.Error(err))
			continue
		}
		accessToken := data.GetValue(oauthAccessTokenVar)
		instanceURL := data.GetValue(instanceURLVar)
		orgID := data.GetValue(orgIDVar)

		h.Subscribe(instanceURL, orgID, accessToken)
	}
}
