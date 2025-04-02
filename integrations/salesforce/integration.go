package salesforce

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
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
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, o *oauth.OAuth, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.SalesforceWebContent)

	h := newHTTPHandler(l, v, o, d)
	common.RegisterSaveHandler(m, desc, h.handleSave)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)

	ctx := authcontext.SetAuthnSystemUser(context.Background())
	h.reopenExistingPubSubConnections(ctx)
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

func (h handler) reopenExistingPubSubConnections(ctx context.Context) {
	cids, err := h.vars.FindConnectionIDs(ctx, desc.ID(), instanceURLVar, "")
	if err != nil {
		h.logger.Error("failed to list connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), orgIDVar, instanceURLVar)
		if err != nil {
			h.logger.Error("can't restart Salesforce PubSub connection",
				zap.String("connection_id", cid.String()), zap.Error(err),
			)
			continue
		}

		cfg, _, err := h.oauth.GetConfig(ctx, desc.UniqueName().String(), cid)
		if err != nil {
			h.logger.Error("failed to get Salesforce OAuth config", zap.Error(err))
			continue
		}

		orgID := vs.GetValue(orgIDVar)
		instanceURL := vs.GetValue(instanceURLVar)

		h.subscribe(cfg.ClientID, orgID, instanceURL, cid)
	}
}
