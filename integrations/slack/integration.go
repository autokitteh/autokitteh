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

	// TODO(INT-167): Remove this after all cloud envs have v0.14.4+ deployed.
	migrateOldConnectionVars(l, v)

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

// TODO(INT-167): Remove this after all cloud envs have v0.14.4+ deployed.
func migrateOldConnectionVars(l *zap.Logger, v sdkservices.Vars) {
	ctx := context.Background()
	iid := sdktypes.NewIntegrationIDFromName(desc.UniqueName().String())
	cids, err := v.FindConnectionIDs(ctx, iid, common.AuthTypeVar, "")
	if err != nil {
		l.Error("failed to list old Slack connection IDs", zap.Error(err))
		return
	}

	for _, cid := range cids {
		l := l.With(zap.String("connection_id", cid.String()))
		l.Info("migrating old Slack connection")
		vsid := sdktypes.NewVarScopeID(cid)

		pairs := []struct{ old, new string }{
			{"Key", "install_ids"},
			{"oauth_TokenTyp", "oauth_TokenType"},

			{"oauth_AccessToken", "oauth_access_token"},
			{"oauth_Expiry", "oauth_expiry"},
			{"oauth_RefreshToken", "oauth_refresh_token"},
			{"oauth_TokenType", "oauth_token_type"},

			{"client_id", "private_client_id"},
			{"client_secret", "private_client_secret"},
			{"signing_secret", "private_signing_secret"},

			{"AppID", "app_id"},
			{"EnterpriseID", "enterprise_id"},
			{"TeamID", "team_id"},

			{"AppToken", "private_app_token"},
			{"BotToken", "private_bot_token"},
		}
		for _, pair := range pairs {
			o, n := sdktypes.NewSymbol(pair.old), sdktypes.NewSymbol(pair.new)
			if err := common.RenameVar(ctx, v, vsid, o, n); err != nil {
				l.Error("failed to migrate Slack connection var name",
					zap.String("old", pair.old),
					zap.String("new", pair.new),
					zap.Error(err),
				)
				continue
			}
		}

		if err := common.MigrateAuthType(ctx, v, vsid); err != nil {
			l.Error("failed to migrate Slack connection's auth type", zap.Error(err))
		}

		if err := common.MigrateDateTimeToRFC3339(ctx, v, vsid, common.OAuthExpiryVar); err != nil {
			l.Error("failed to migrate Slack connection's OAuth expiry", zap.Error(err))
		}

		_ = v.Delete(ctx, vsid, sdktypes.NewSymbol("authType"))
	}
}
