package microsoft

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/connection"
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

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
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
	// Use the AutoKitteh server's default Microsoft OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// First save the user-provided details of a private Microsoft OAuth 2.0 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthPrivate:
		if err := h.savePrivateOAuthApp(r, vsid); err != nil {
			c.AbortBadRequest(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	// Same as a private OAuth 2.0 app, but without the OAuth 2.0 flow
	// (it uses application permissions instead of user-delegated ones).
	case integrations.DaemonApp:
		if err := h.savePrivateDaemonApp(r, c.Integration, cid); err != nil {
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

// savePrivateOAuthApp saves the user-provided details of a
// private Microsoft OAuth 2.0 app as connection variables.
func (h handler) savePrivateOAuthApp(r *http.Request, vsid sdktypes.VarScopeID) error {
	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		tenantID = "common"
	}

	app := connection.PrivateApp{
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		TenantID:     tenantID,
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.ClientID == "" || (app.ClientSecret == "" && app.Certificate == "") {
		return errors.New("missing private app details")
	}

	return h.vars.Set(r.Context(), sdktypes.EncodeVars(app).WithScopeID(vsid)...)
}

// savePrivateDaemonApp saves the user-provided details of
// a private Microsoft Daemon app as connection variables.
func (h handler) savePrivateDaemonApp(r *http.Request, i sdktypes.Integration, cid sdktypes.ConnectionID) error {
	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		tenantID = "common"
	}

	app := connection.PrivateApp{
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		TenantID:     tenantID,
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.ClientID == "" || (app.ClientSecret == "" && app.Certificate == "") {
		return errors.New("missing private app details")
	}

	// Test the app's usability by generating a new token.
	ctx := r.Context()
	vsid := sdktypes.NewVarScopeID(cid)
	vs, err := h.vars.Get(ctx, vsid)
	if err != nil {
		h.logger.Error("failed to read connection vars", zap.Error(err))
		return errors.New("failed to read connection vars")
	}

	t, err := connection.DaemonToken(ctx, vs)
	if err != nil {
		h.logger.Error("failed to generate MS daemon app token", zap.Error(err))
		return err
	}

	vs = sdktypes.EncodeVars(app)

	// Optional: save the tenant details, if the app is allowed to read them.
	if org, err := connection.GetOrgInfo(ctx, t); err == nil {
		vs = vs.Append(sdktypes.EncodeVars(org)...)
	}

	// Subscribe to receive asynchronous change notifications from
	// Microsoft Graph, based on the connection's integration type.
	svc := connection.NewServices(h.logger, h.vars, h.oauth)
	err = errors.Join(connection.Subscribe(ctx, svc, cid, resources(i))...)
	if err != nil {
		h.logger.Error("failed to create MS event subscriptions", zap.Error(err))
		return err
	}

	return h.vars.Set(ctx, vs.WithScopeID(vsid)...)
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger) {
	urlPath, err := oauthURL(c.ConnectionID, c.Origin, r.FormValue("auth_scopes"))
	if err != nil {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
		return
	}
	http.Redirect(w, r, urlPath, http.StatusFound)
}

// oauthURL constructs a relative URL to start a 3-legged OAuth
// 2.0 flow, using the AutoKitteh server's generic OAuth service.
func oauthURL(cid, origin, scopes string) (string, error) {
	// Security check: parameters must be alphanumeric strings,
	// to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^\w+$`)
	if !re.MatchString(cid + origin + scopes) {
		return "", errors.New("invalid connection ID, origin, or scopes")
	}

	// Default scopes: all ("microsoft").
	// Narrowed-down scopes: "microft-excel", "microsoft-teams", etc.
	path := "/oauth/start/microsoft"
	if scopes != "" {
		path += "_" + scopes
	}

	// Remember the AutoKitteh connection ID and request origin.
	return fmt.Sprintf("%s?cid=%s&origin=%s", path, cid, origin), nil
}
