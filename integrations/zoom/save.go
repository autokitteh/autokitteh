package zoom

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
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
	authType := h.saveAuthType(r.Context(), vsid, r.FormValue("auth_type"))
	l = l.With(zap.String("auth_type", authType))

	switch authType {
	// Use the AutoKitteh server's default Zoom OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// First save the user-provided details of a private Zoom OAuth 2.0 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthPrivate:
		app, err := h.savePrivateApp(r, vsid)
		if err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		if app.ClientID == "" || app.ClientSecret == "" {
			l.Error("save connection: missing private app details")
			c.AbortBadRequest("missing private app details")
			return
		}
		startOAuth(w, r, c, l)

	// Same as a private OAuth 2.0 app, but without the OAuth 2.0 flow
	// (it uses application permissions instead of user-delegated ones).
	case integrations.ServerToServer:
		app, err := h.savePrivateApp(r, vsid)
		if err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		if app.AccountID == "" || app.ClientID == "" || app.ClientSecret == "" {
			l.Error("save connection: missing private app details")
			c.AbortBadRequest("missing private app details")
			return
		}
		if _, err := serverToken(r.Context(), app); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		urlPath, err := c.FinalURL()
		if err != nil {
			l.Error("save connection: failed to construct final URL", zap.Error(err))
			c.AbortServerError("bad redirect URL")
			return
		}
		http.Redirect(w, r, urlPath, http.StatusFound)

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type")
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", authType))
	}
}

// saveAuthType saves the authentication type that the user selected for this connection.
// This will be redundant if/when the only way to initialize connections is via the web UI.
// Therefore, we do not care if this function fails to save it as a connection variable.
func (h handler) saveAuthType(ctx context.Context, vsid sdktypes.VarScopeID, authType string) string {
	v := sdktypes.NewVar(authTypeVar).SetValue(authType)
	_ = h.vars.Set(ctx, v.WithScopeID(vsid))
	return authType
}

// savePrivateApp saves the user-provided details of a private Zoom
// OAuth 2.0 app or Service-to-Service internal app as connection variables.
func (h handler) savePrivateApp(r *http.Request, vsid sdktypes.VarScopeID) (*privateApp, error) {
	app := privateApp{
		AccountID:    r.FormValue("account_id"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		SecretToken:  r.FormValue("secret_token"),
	}

	if err := h.vars.Set(r.Context(), sdktypes.EncodeVars(app).WithScopeID(vsid)...); err != nil {
		return nil, err
	}

	return &app, nil
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

	urlPath := fmt.Sprintf("/oauth/start/zoom?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, urlPath, http.StatusFound)
}
