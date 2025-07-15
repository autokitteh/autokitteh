package airtable

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

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
	// Use the AutoKitteh server's default Zoom OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		h.startOAuth(w, r, c, l)

	// Save the user-provided personal access token (PAT) and finish.
	case integrations.PAT:
		if err := h.savePAT(r, vsid); err != nil {
			l.Error("save connection: " + err.Error())
			c.AbortBadRequest(err.Error())
			return
		}

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type")
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", authType))
	}
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func (h handler) startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger) {
	// Security check: parameters must be alphanumeric strings,
	// to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^\w+$`)
	if !re.MatchString(c.ConnectionID + c.Origin) {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
		return
	}

	urlPath := fmt.Sprintf("/oauth/start/airtable?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, urlPath, http.StatusFound)
}

func (h handler) savePAT(r *http.Request, vsid sdktypes.VarScopeID) error {
	// Parse the PAT from the request.
	pat := r.FormValue("pat")
	if pat == "" {
		return errors.New("missing personal access token (PAT)")
	}

	// Save the PAT as a connection variable.
	v := sdktypes.NewVar(common.PATVar).SetValue(pat).SetSecret(true)
	return h.vars.Set(r.Context(), v.WithScopeID(vsid))
}
