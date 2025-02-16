package slack

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/slack/vars"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/integrations/slack/websockets"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

var desc = common.Descriptor("slack", "Slack", "/static/images/slack.svg")

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
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.SlackWebContent)

	iid := sdktypes.NewIntegrationIDFromName(desc.UniqueName().String())
	wsh := websockets.NewHandler(l, v, d, iid)

	h := newHTTPHandler(l, v, d, wsh)
	common.RegisterSaveHandler(m, desc, h.handleSave)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)

	// TODO(INT-167): Remove "custom-oauth" once the web UI is migrated too.
	pattern := " /slack/custom-oauth/save"
	m.Auth.HandleFunc(http.MethodGet+pattern, h.handleSave)
	m.Auth.HandleFunc(http.MethodPost+pattern, h.handleSave)

	// Event webhooks (no AutoKitteh user authentication by definition, because
	// these asynchronous requests are sent to us by third-party services).
	whh := webhooks.NewHandler(l, v, d, iid)
	m.NoAuth.HandleFunc("POST "+webhooks.BotEventPath, whh.HandleBotEvent)
	m.NoAuth.HandleFunc("POST "+webhooks.SlashCommandPath, whh.HandleSlashCommand)
	m.NoAuth.HandleFunc("POST "+webhooks.InteractionPath, whh.HandleInteraction)

	reopenExistingWebSocketConnections(context.Background(), l, v, wsh)
}

// handler implements several HTTP webhooks to save authentication data, as
// well as receive and dispatch third-party asynchronous event notifications.
type handler struct {
	logger     *zap.Logger
	vars       sdkservices.Vars
	dispatch   sdkservices.DispatchFunc
	webSockets websockets.Handler
}

func newHTTPHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.DispatchFunc, h websockets.Handler) handler {
	return handler{logger: l, vars: v, dispatch: d, webSockets: h}
}

// reopenExistingWebSocketConnections initializes a new WebSocket pool
// for existing Socket Mode connections when the AutoKitteh server starts.
func reopenExistingWebSocketConnections(ctx context.Context, l *zap.Logger, v sdkservices.Vars, h websockets.Handler) {
	iid := sdktypes.NewIntegrationIDFromName(desc.UniqueName().String())
	cids, err := v.FindConnectionIDs(ctx, iid, vars.AppTokenVar, "")
	if err != nil {
		l.Error("failed to list WebSocket-based connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		data, err := v.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("can't restart Slack Socket Mode connection",
				zap.String("connection_id", cid.String()),
				zap.Error(err),
			)
			continue
		}

		appID := data.GetValue(vars.AppIDVar)
		appToken := data.GetValue(vars.AppTokenVar)
		botToken := data.GetValue(vars.BotTokenVar)
		h.OpenWebSocketConnection(appID, appToken, botToken)
	}
}
