package linear

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
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
	// Use the AutoKitteh server's default Linear OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		if err := h.saveActor(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	// First save the user-provided details of a private Linear OAuth 2.0 app,
	// and only then redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthPrivate:
		if err := h.saveActor(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortServerError(err.Error())
			return
		}
		if err := h.savePrivateOAuth(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortBadRequest(err.Error())
			return
		}
		startOAuth(w, r, c, l)

	// Check and save user-provided details, no 3-legged OAuth 2.0 flow is needed.
	case integrations.APIKey:
		if err := h.saveAPIKey(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
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

// saveActor saves Linear's OAuth "actor" (user/app) parameter as a connection variable.
func (h handler) saveActor(r *http.Request, vsid sdktypes.VarScopeID) error {
	v := sdktypes.NewVar(actorVar).SetValue(r.FormValue("actor"))
	return h.vars.Set(r.Context(), v.WithScopeID(vsid))
}

// savePrivateOAuth saves the user-provided details of a
// private Linear OAuth 2.0 app as connection variables.
func (h handler) savePrivateOAuth(r *http.Request, vsid sdktypes.VarScopeID) error {
	app := privateOAuth{
		ClientID:      r.FormValue("client_id"),
		ClientSecret:  r.FormValue("client_secret"),
		WebhookSecret: r.FormValue("webhook_secret"),
	}

	// Sanity check: all the required details were provided, and are valid.
	if app.ClientID == "" || app.ClientSecret == "" {
		return errors.New("missing private OAuth 2.0 details")
	}

	return h.vars.Set(r.Context(), sdktypes.EncodeVars(app).WithScopeID(vsid)...)
}

// saveAPIKey saves a user-provided API key as a connection variable.
func (h handler) saveAPIKey(r *http.Request, vsid sdktypes.VarScopeID) error {
	apiKey := r.FormValue("api_key")
	if apiKey == "" {
		return errors.New("missing API key")
	}

	// Test the API key's usability and get authoritative connection details.
	ctx := r.Context()
	org, viewer, err := orgAndViewerInfo(ctx, apiKey)
	if err != nil {
		return errors.New("API key test failed")
	}

	vs := sdktypes.NewVars(sdktypes.NewVar(apiKeyVar).SetValue(apiKey).SetSecret(true))
	vs = vs.Append(sdktypes.EncodeVars(org)...)
	vs = vs.Append(sdktypes.EncodeVars(viewer)...)
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

	urlPath := fmt.Sprintf("/oauth/start/linear?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, urlPath, http.StatusFound)
}
