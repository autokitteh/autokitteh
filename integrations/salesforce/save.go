package salesforce

import (
	"context"
	"errors"
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

	switch authType {
	case integrations.OAuthPrivate:
		if err := h.savePrivateOAuth(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	default:
		l.Error("save connection: unknown auth type", zap.String("auth_type", authType))
		c.AbortBadRequest("unknown auth type")
	}
}

// saveAuthType saves the auth type for a connection.
func (h handler) saveAuthType(ctx context.Context, vsid sdktypes.VarScopeID, authType string) string {
	v := sdktypes.NewVar(authTypeVar).SetValue(authType)
	_ = h.vars.Set(ctx, v.WithScopeID(vsid))
	return authType
}

// savePrivateOAuth saves the user-provided details of a
// private Salesforce OAuth 2.0 app as connection variables.
func (h handler) savePrivateOAuth(r *http.Request, vsid sdktypes.VarScopeID) error {
	app := privateOAuth{
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.ClientID == "" || app.ClientSecret == "" {
		return errors.New("missing private OAuth 2.0 details")
	}

	return h.vars.Set(r.Context(), sdktypes.EncodeVars(app).WithScopeID(vsid)...)
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

	urlPath := fmt.Sprintf("/oauth/start/salesforce?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, urlPath, http.StatusFound)
}
