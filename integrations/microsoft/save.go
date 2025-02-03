package microsoft

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
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
	authType := h.saveAuthType(r.Context(), sdktypes.NewVarScopeID(cid), r.FormValue("auth_type"))

	switch authType {
	// Use the AutoKitteh's server's default Microsoft OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// First save the user-provided details of a private Microsoft OAuth 2.0 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthPrivate:
		if err := h.savePrivateAppConfig(r, cid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	// Same as a private OAuth 2.0 app, but without the OAuth 2.0 flow
	// (it uses application permissions instead of user-delegated ones).
	case integrations.DaemonApp:
		if err := h.savePrivateAppConfig(r, cid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		urlPath, err := c.FinalURL()
		if err != nil {
			l.Error("failed to construct final OAuth URL", zap.Error(err))
			c.AbortServerError("save connection: bad redirect URL")
			return
		}
		http.Redirect(w, r, urlPath, http.StatusFound)

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type", zap.String("auth_type", authType))
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", authType))
	}
}

// saveAuthType saves the authentication type that the user selected for this connection.
// This will be redundant if/when the only way to initialize connections is via the web UI.
// Therefore, we do not care if this function fails to save it as a connection variable.
func (h handler) saveAuthType(ctx context.Context, vsid sdktypes.VarScopeID, authType string) string {
	v := sdktypes.NewVar(connection.AuthTypeVar).WithScopeID(vsid)
	_ = h.vars.Set(ctx, v.SetValue(authType))
	return authType
}

// savePrivateAppConfig saves the user-provided details of
// a private Microsoft OAuth 2.0 app as connection variables.
func (h handler) savePrivateAppConfig(r *http.Request, cid sdktypes.ConnectionID) error {
	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		tenantID = "common"
	}

	app := connection.PrivateAppConfig{
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		TenantID:     tenantID,
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.ClientID == "" || (app.ClientSecret == "" && app.Certificate == "") {
		return errors.New("missing private app details")
	}

	ctx := r.Context()
	vsid := sdktypes.NewVarScopeID(cid)
	vs, err := h.vars.Get(ctx, vsid)
	if err != nil {
		return fmt.Errorf("failed to read connection vars: %w", err)
	}

	t, err := connection.DaemonToken(ctx, vs)
	if err != nil {
		return err
	}
	vs = sdktypes.EncodeVars(app)

	// Optional: save the tenant details, if the app is allowed to read them.
	if org, err := connection.GetOrgInfo(ctx, t); err == nil {
		vs = vs.Append(sdktypes.EncodeVars(org)...)
	}

	if err := h.vars.Set(r.Context(), vs.WithScopeID(vsid)...); err != nil {
		return err
	}

	// https://learn.microsoft.com/en-us/graph/teams-change-notification-in-microsoft-teams-overview
	resources := []string{
		"/chats",
		"/chats/getAllMembers",
		"/chats/getAllMessages",
		"/teams",
		"/teams/getAllChannels",
		"/teams/getAllMessages",
	}

	svc := connection.NewServices(h.logger, h.vars, h.oauth)
	var errs []error
	for _, r := range resources {
		if err := connection.CreateSubscription(ctx, svc, cid, r); err != nil {
			errs = append(errs, err)
		}
	}
	// TODO(INT-203): "Subscription operations for tenant-wide chats subscription is not allowed in 'OnBehalfOfUser' context."
	if len(errs) > 0 {
		h.logger.Error("failed to create event subscriptions", zap.Errors("errors", errs))
		// return errors.Join(errs...)
	}

	return nil
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger) {
	urlPath, err := oauthURL(c.ConnectionID, c.Origin, r.FormValue("auth_scopes"))
	if err != nil {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
	}
	http.Redirect(w, r, urlPath, http.StatusFound)
}

// oauthURL constructs a relative URL to start a 3-legged OAuth
// 2.0 flow, using the AutoKitteh server's generic OAuth service.
func oauthURL(cid, origin, scopes string) (string, error) {
	// Security check: parameters must be alphanumeric strings,
	// to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^[\w]+$`)
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
	return path + fmt.Sprintf("?cid=%s&origin=%s", cid, origin), nil
}
