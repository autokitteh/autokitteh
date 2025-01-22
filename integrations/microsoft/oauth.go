package microsoft

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	authTypeVar = sdktypes.NewSymbol("auth_type")

	clientIDVar     = sdktypes.NewSymbol("client_id")
	clientSecretVar = sdktypes.NewSymbol("client_secret")
)

// isSecrets shows which connection variables are secrets.
func isSecret(varName sdktypes.Symbol) bool {
	return varName == clientSecretVar
}

// handleOAuth receives an incoming redirect request from AutoKitteh's
// generic OAuth service, which contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation.
// This is the last step in a 3-legged OAuth 2.0 flow, in which we verify
// the usability of the OAuth token, and save it as connection variables.
func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Parse the GET request's query params.
	if err := r.ParseForm(); err != nil {
		l.Warn("save connection after OAuth flow: failed to parse HTTP request", zap.Error(err))
		c.AbortBadRequest("OAuth redirection parsing error")
		return
	}

	// Handle OAuth errors (e.g. the user didn't authorize us), based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirection reported an error", zap.Error(errors.New(e)))
		c.AbortBadRequest(e)
		return
	}

	// TODO(BEFORE MERGE): Finish implementation.
}
