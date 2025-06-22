package airtable

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var pkce_verifier = sdktypes.NewSymbol("pkce_verifier")

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
		h.startOAuth(w, r, c, l, vsid)

	// Save the user-provided personal access token (PAT) and finish.
	case integrations.PAT:
		// TODO: Implement PAT support.

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type")
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", authType))
	}
}

// PKCE requires a code_verifier between 43–128 characters from [A-Z, a-z, 0-9, "-", ".", "_", "~"]
func generateCodeVerifier() (string, error) {
	// 64 bytes gives ~86-character base64 string
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	return verifier, nil
}

func codeChallengeS256(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func (h handler) startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger, vsid sdktypes.VarScopeID) {
	// Security check: parameters must be alphanumeric strings,
	// to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^\w+$`)
	if !re.MatchString(c.ConnectionID + c.Origin) {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
		return
	}

	verifier, err := generateCodeVerifier()
	if err != nil {
		log.Fatal(err)
	}

	vs := sdktypes.NewVars(sdktypes.NewVar(pkce_verifier).SetValue(verifier).SetSecret(true))
	if err := h.vars.Set(r.Context(), vs.WithScopeID(vsid)...); err != nil {
		l.Warn("failed to save vars", zap.Error(err))
		c.AbortServerError("failed to save connection variables")
	}

	challenge := codeChallengeS256(verifier)

	urlPath := fmt.Sprintf("/oauth/start/airtable?cid=%s&origin=%s&code_challenge=%s&code_challenge_method=S256", c.ConnectionID, c.Origin, challenge)
	http.Redirect(w, r, urlPath, http.StatusFound)
}
