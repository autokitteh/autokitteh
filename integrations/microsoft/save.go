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
	vsid := sdktypes.NewVarScopeID(cid)

	// Determine what to save and how to proceed.
	authType := h.saveAuthType(r.Context(), vsid, r.FormValue("auth_type"))

	switch authType {
	// Use the AutoKitteh's server's default Microsoft OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// First save the user-provided details of a private Microsoft OAuth 2.0 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthPrivate:
		if err := h.saveOAuthAppConfig(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		startOAuth(w, r, c, l)

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

// OAuthAppConfig contains the user-provided details of a private Microsoft OAuth 2.0 app.
type OAuthAppConfig struct {
	ClientID     string `var:"client_id"`
	ClientSecret string `var:"client_secret,secret"`
	Tenant       string `var:"tenant"`
}

// saveOAuthAppConfig saves the user-provided details of a
// private Microsoft OAuth 2.0 app as connection variables.
func (h handler) saveOAuthAppConfig(r *http.Request, vsid sdktypes.VarScopeID) error {
	tenant := r.FormValue("tenant")
	if tenant == "" {
		tenant = "common"
	}

	app := OAuthAppConfig{
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		Tenant:       tenant,
	}

	// Sanity check: all the required details were provided.
	if app.ClientID == "" || app.ClientSecret == "" {
		return errors.New("missing OAuth 2.0 app details")
	}

	return h.vars.Set(r.Context(), sdktypes.EncodeVars(app).WithScopeID(vsid)...)
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
		path += "-" + scopes
	}

	// Remember the AutoKitteh connection ID and request origin.
	return path + fmt.Sprintf("?cid=%s&origin=%s", cid, origin), nil
}
