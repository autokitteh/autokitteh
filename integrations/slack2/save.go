package slack2

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack2/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleSave saves connection variables for an AutoKitteh connection.
// This may result in a fully-initialized and usable connection, or it
// may be an intermediate step before starting a 3-legged OAuth 2.0 flow.
// This handler accepts both GET and POST requests alike. Why GET? This
// is the only option when the web UI opens a pop-up window for OAuth.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header in POST requests.
	contentType := r.Header.Get("Content-Type")
	expected := "application/x-www-form-urlencoded"
	if r.Method == http.MethodPost && !strings.HasPrefix(contentType, expected) {
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", contentType))
		c.AbortBadRequest("unexpected request content type")
		return
	}

	// Parse GET request's query params / POST request's body.
	if err := r.ParseForm(); err != nil {
		l.Warn("save connection: failed to parse HTTP request", zap.Error(err))
		c.AbortBadRequest("request parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	// Determine what to save and how to proceed.
	vsid := sdktypes.NewVarScopeID(cid)
	authType := common.SaveAuthType(r, h.vars, vsid)
	l = l.With(zap.String("auth_type", authType))

	switch authType {
	// Use the AutoKitteh server's default Slack OAuth v2 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	// TODO(INT-267): Remove [integrations.OAuth] once the web UI is migrated too.
	case integrations.OAuth, integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// First save the user-provided details of a private Slack OAuth v2 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthPrivate:
		if err := h.savePrivateOAuth(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortBadRequest(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	// Check and save user-provided details, no 3-legged OAuth 2.0 flow is needed.
	case integrations.SocketMode:
		if err := h.saveSocketModeApp(r, vsid); err != nil {
			c.AbortBadRequest(err.Error())
			return
		}
		urlPath, err := c.FinalURL()
		if err != nil {
			l.Error("save connection: failed to construct final URL", zap.Error(err))
			c.AbortBadRequest("bad redirect URL")
			return
		}
		http.Redirect(w, r, urlPath, http.StatusFound)

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type")
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", authType))
	}
}

// savePrivateOAuth saves the user-provided details of
// a private Slack OAuth v2 app as connection variables.
func (h handler) savePrivateOAuth(r *http.Request, vsid sdktypes.VarScopeID) error {
	app := vars.PrivateOAuth{
		ClientID:      r.FormValue("client_id"),
		ClientSecret:  r.FormValue("client_secret"),
		SigningSecret: r.FormValue("signing_secret"),
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.ClientID == "" || app.ClientSecret == "" || app.SigningSecret == "" {
		return errors.New("missing private OAuth 2.0 app details")
	}

	return h.vars.Set(r.Context(), sdktypes.EncodeVars(app).WithScopeID(vsid)...)
}

// saveSocketModeApp saves the user-provided details of
// a private Slack Socket Mode app as connection variables.
func (h handler) saveSocketModeApp(r *http.Request, vsid sdktypes.VarScopeID) error {
	app := vars.SocketMode{
		BotToken: r.FormValue("bot_token"),
		AppToken: r.FormValue("app_token"),
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.BotToken == "" || app.AppToken == "" {
		return errors.New("missing private Socket Mode app details")
	}

	// Test the tokens' usability and get authoritative installation details.
	ctx := r.Context()
	auth, err := api.AuthTest(ctx, app.BotToken)
	if err != nil {
		h.logger.Warn("Slack token auth test failed", zap.Error(err))
		return errors.New("Slack token auth test failed")
	}

	bot, err := api.BotsInfo(ctx, app.BotToken, auth)
	if err != nil {
		h.logger.Warn("Slack bot info request failed", zap.Error(err))
		return errors.New("Slack bot info request failed")
	}

	_, err = api.AppsConnectionsOpen(ctx, app.AppToken)
	if err != nil {
		h.logger.Warn("Slack WebSocket connection opening failed", zap.Error(err))
		return errors.New("Slack WebSocket connection opening failed")
	}

	// Open a new Socket Mode connection.
	h.webSockets.OpenWebSocketConnection(bot.AppID, app.AppToken, app.BotToken)

	vs := sdktypes.EncodeVars(app)
	vs = vs.Append(sdktypes.EncodeVars(encodeInstallInfo(auth, bot))...)
	return h.vars.Set(r.Context(), vs.WithScopeID(vsid)...)
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger) {
	// Security check: parameters must be alphanumeric strings,
	// to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^\w+$`)
	if !re.MatchString(c.ConnectionID + c.Origin) {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
		return
	}

	urlPath := fmt.Sprintf("/oauth/start/slack?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, urlPath, http.StatusFound)
}

// encodeInstallInfo encodes a Slack app's installation into AutoKitteh connection variables.
func encodeInstallInfo(auth *api.AuthTestResponse, bot *api.Bot) vars.InstallInfo {
	return vars.InstallInfo{
		EnterpriseID: auth.EnterpriseID,
		Team:         auth.Team,
		TeamID:       auth.TeamID,
		User:         auth.User,
		UserID:       auth.UserID,

		BotName:    bot.Name,
		BotID:      bot.ID,
		BotUpdated: time.Unix(int64(bot.Updated), 0).UTC().Format(time.RFC3339),
		AppID:      bot.AppID,

		InstallIDs: vars.InstallIDs(bot.AppID, auth.EnterpriseID, auth.TeamID),
	}
}
