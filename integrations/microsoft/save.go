package microsoft

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

	// Determine what to save and how to proceed.
	switch at := r.FormValue("auth_type"); at {
	// Use the AutoKitteh's server's default Microsoft OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// First save the user-provided details of a custom Microsoft OAuth 2.0 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthCustom:
		if err := h.saveOAuthAppConfig(r, c); err != nil {
			l.Warn("save connection: " + err.Error())
			c.AbortBadRequest(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type", zap.String("auth_type", at))
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", at))
	}
}

// saveOAuthAppConfig saves the user-provided details of a
// custom Microsoft OAuth 2.0 app as connection variables.
func (h handler) saveOAuthAppConfig(r *http.Request, c sdkintegrations.ConnectionInit) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	m := map[sdktypes.Symbol]string{
		clientIDVar:     r.FormValue("client_id"),
		clientSecretVar: r.FormValue("client_secret"),
	}

	// Sanity check: all the required details were provided.
	if m[clientIDVar] == "" || m[clientSecretVar] == "" {
		return errors.New("missing OAuth 2.0 app details")
	}

	return h.vars.Set(r.Context(), mapToVars(m, sdktypes.NewVarScopeID(cid))...)
}

// mapToVars converts a map of key-value pairs to a Vars object for a given connection.
func mapToVars(m map[sdktypes.Symbol]string, vsid sdktypes.VarScopeID) sdktypes.Vars {
	vs := sdktypes.NewVars()
	for name, val := range m {
		v := sdktypes.NewVar(name).WithScopeID(vsid).SetValue(val)
		vs = vs.Append(v.SetSecret(isSecret(name)))
	}
	return vs
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger) {
	if urlPath := oauthURL(r.Form, c); urlPath != "" {
		http.Redirect(w, r, urlPath, http.StatusFound)
	} else {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
	}
}

// oauthURL constructs a relative URL to start a 3-legged OAuth
// 2.0 flow, using the AutoKitteh server's generic OAuth service.
func oauthURL(vs url.Values, c sdkintegrations.ConnectionInit) string {
	// Default scopes: all ("microsoft").
	path := "/oauth/start/microsoft-%s?cid=%s&origin=%s"

	// Narrow down the requested scopes to a specific Microsoft 365 product?
	scopes := vs.Get("auth_scopes")

	// Remember the AutoKitteh connection ID and request origin.
	path = fmt.Sprintf(path, scopes, c.ConnectionID, c.Origin)
	path = strings.ReplaceAll(path, "-?", "?")

	// Security checks: ensure the URL is relative
	// and doesn't contain suspicious characters.
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return ""
	}
	u, err := url.Parse(path)
	if err != nil || u.IsAbs() || u.Hostname() != "" {
		return ""
	}

	return path
}
