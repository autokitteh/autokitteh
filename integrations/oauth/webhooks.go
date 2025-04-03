package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// startOAuthFlow starts a 3-legged OAuth 2.0 flow by redirecting the user (via a web
// page) to the authorization endpoint of a third-party service named in the request.
func (o *OAuth) startOAuthFlow(w http.ResponseWriter, r *http.Request) {
	integ := r.PathValue("integration")
	id := r.FormValue("cid")
	origin := r.FormValue("origin")

	l := o.logger.With(
		zap.String("url_path", r.URL.Path),
		zap.String("integration", integ),
		zap.String("connection_id", id),
		zap.String("origin", origin),
	)

	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		l.Warn("failed to parse connection ID", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	cfg, opts, err := o.GetConfig(r.Context(), integ, cid)
	if err != nil {
		l.Warn("failed to get OAuth config", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// The origin is needed at the end of the flow, to report success/errors correctly.
	if origin == "" {
		l.Warn("missing origin parameter")
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Identify the relevant connection when we get an OAuth response.
	state := strings.Replace(id, "con_", "", 1) + "_" + origin

	u := cfg.AuthCodeURL(state, authCodes(opts)...)
	http.Redirect(w, r, u, http.StatusFound)
}

// exchangeCodeToToken receives a redirect back from a third-party service's
// authorization endpoint (the OAuth 2.0 2nd leg), and exchanges the received
// authorization code for an new access token (the 3rd leg). If all goes well,
// it redirects the token back to the named integration's own OAuth webhook in
// order to complete the initialization procedure of an AutoKitteh connection.
func (o *OAuth) exchangeCodeToToken(w http.ResponseWriter, r *http.Request) {
	integ := r.PathValue("integration")
	state := r.FormValue("state")

	// This webhook is unauthenticated by definition (3rd-party services redirect to it),
	// so we need to elevate the webhook's permission to read from AutoKitteh's database.
	// This is secure because only trusted OAuth providers with preconfigured apps can
	// generate valid code parameters related to valid connection IDs: fake, replayed,
	// or unrelated parameters are checked and rejected by AutoKitteh as well as the
	// 3rd-party OAuth provider. On top of that, the elevated permission cannot be
	// hijacked: it's used only for reading, and the results are never exposed.
	ctx := authcontext.SetAuthnSystemUser(r.Context())

	l := o.logger.With(
		zap.String("url_path", r.URL.Path),
		zap.String("integration", integ),
		zap.String("state", state),
	)

	// An unrecognized integration name is a critical error
	// (the OAuth app's callback URL was configured incorrectly).
	_, ok := o.oauthConfigs[integ]
	if !ok {
		l.Error("unrecognized integration name")
		http.Error(w, "Unrecognized integration name in callback URL path", http.StatusBadRequest)
		return
	}

	// Check for reported errors before trying to parse and use the state parameter
	// (so we don't lose them even if the state is missing or malformed).
	oauthErrParam := r.FormValue("error_description")
	if oauthErrParam == "" {
		oauthErrParam = r.FormValue("error")
	}
	if oauthErrParam != "" {
		l = l.With(zap.String("error_param", oauthErrParam))
	}

	// If there is no error, but the state parameter is missing: the OAuth
	// flow was initiated outside of AutoKitteh. We can't continue without
	// a connection ID and origin, so for the sake of a good user experience
	// we redirect to the landing page of this AutoKitteh server's frontend.
	if oauthErrParam == "" && state == "" {
		http.Redirect(w, r, guessFrontendURL(o.BaseURL), http.StatusFound)
		return
	}

	// At this point, the state parameter must exist, but it may be malformed.
	cid, origin, err := parseStateParam(state)
	if err != nil {
		l.Warn(err.Error())
		// The state parameter is malformed, but if an OAuth error was
		// reported above, it takes precedence over the state parameter's
		// error, in terms of the error message we show to the user.
		errMsg := err.Error()
		if oauthErrParam != "" {
			errMsg = oauthErrParam
		}
		abort(w, r, cid, integ, origin, errMsg)
		return
	}

	// Report back OAuth errors, if any were detected above (the state
	// parameter is well-formed, but the OAuth flow has failed).
	if oauthErrParam != "" {
		abort(w, r, cid, integ, origin, oauthErrParam)
		return
	}

	// Special case: we already have what we need to generate JWTs for GitHub connections
	// (i.e. the GitHub app's installation ID), no need to exchange the OAuth code.
	if o.flags(integ).useJWTsNotOAuth {
		l = l.With(
			zap.String("github_setup_action", r.FormValue("setup_action")),
			zap.String("github_installation_id", r.FormValue("installation_id")),
		)

		data, err := sdkintegrations.OAuthData{Params: r.URL.Query()}.Encode()
		if err != nil {
			l.Error("OAuth URL parameters encoding error", zap.Error(err))
			abort(w, r, cid, integ, origin, "OAuth URL parameters encoding error")
			return
		}

		l.Info("successful GitHub app flow")
		redirect(w, r, cid, integ, origin, "oauth", data)
		return
	}

	// Otherwise, we need to exchange the received authorization code for an access token,
	// potentially including a refresh token (this is the 3rd leg of the OAuth 2.0 flow).
	code := r.FormValue("code")
	if code == "" {
		l.Warn("missing code parameter in OAuth redirect",
			zap.String("raw_url", r.RequestURI),
			zap.Any("query", r.URL.Query()),
		)
		abort(w, r, cid, integ, origin, "missing code parameter in OAuth redirect")
		return
	}

	cfg, opts, err := o.GetConfig(ctx, integ, cid)
	if err != nil {
		l.Warn("failed to get OAuth config", zap.Error(err))
		abort(w, r, cid, integ, origin, http.StatusText(http.StatusInternalServerError))
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	token, err := cfg.Exchange(ctx, code, authCodes(opts)...)
	if err != nil {
		l.Warn("OAuth code exchange error", zap.Error(err))
		abort(w, r, cid, integ, origin, "OAuth code exchange error")
		return
	}

	l.Info("successful OAuth token exchange")

	// Finally, pass the OAuth data back to the originating integration.
	data, err := sdkintegrations.OAuthData{Token: token, Params: r.URL.Query(), Extra: extraData(token)}.Encode()
	if err != nil {
		l.Error("OAuth token encoding error", zap.Error(err))
		abort(w, r, cid, integ, origin, "OAuth token encoding error")
		return
	}

	redirect(w, r, cid, integ, origin, "oauth", data)
}

func authCodes(opts map[string]string) []oauth2.AuthCodeOption {
	var acos []oauth2.AuthCodeOption
	for k, v := range opts {
		acos = append(acos, oauth2.SetAuthURLParam(k, v))
	}
	return acos
}

// guessFrontendURL converts the AutoKitteh server's configured
// public URL for webhooks into the URL of the frontend application.
// This is used as a final destination after a successful OAuth flow,
// when the OAuth flow was initiated outside of AutoKitteh.
func guessFrontendURL(backendBaseURL string) string {
	// Multi-tenant cloud: "api.autokitteh.cloud" -> "app.autokitteh.cloud"
	if strings.Contains(backendBaseURL, "//api.") {
		return strings.Replace(backendBaseURL, "//api.", "//app.", 1)
	}

	// Single-tenant cloud: "customer-api.autokitteh.cloud" -> "customer.autokitteh.cloud"
	if strings.Contains(backendBaseURL, "-api.") {
		return strings.Replace(backendBaseURL, "-api.", ".", 1)
	}

	// Probably a self-hosted server where only the backend (port 9980) is exposed
	// and the frontend (port 9982) is not. We could try "http://localhost:9982"
	// but it's a better user experience to redirect to the public cloud.
	return "https://app.autokitteh.cloud"
}

func parseStateParam(state string) (sdktypes.ConnectionID, string, error) {
	sub := regexp.MustCompile(`^([0-9a-z]{26})_([a-z]+)$`).FindStringSubmatch(state)
	if len(sub) != 3 {
		return sdktypes.InvalidConnectionID, "", errors.New("invalid state parameter")
	}

	cid, err := sdktypes.StrictParseConnectionID("con_" + sub[1])
	if err != nil {
		err = fmt.Errorf("invalid connection ID in state parameter: %w", err)
		return sdktypes.InvalidConnectionID, "", err
	}

	return cid, sub[2], nil
}

// redirect forwards the client after the 3rd leg of an OAuth 2.0 flow
// (exchanging an authorization code for a token) to the integration's
// own OAuth webhook, where the connection initialization is completed.
// The "param" and "value" are URL-encoded query parameters, where "param"
// may be "oauth" (in case of success), or "error" (in case of failure).
func redirect(w http.ResponseWriter, r *http.Request, cid sdktypes.ConnectionID, integ, origin, param, value string) {
	u := fmt.Sprintf("/%s/oauth?cid=%s&origin=%s&%s=%s", integ, cid.String(), origin, param, value)
	http.Redirect(w, r, u, http.StatusFound)
}

func abort(w http.ResponseWriter, r *http.Request, cid sdktypes.ConnectionID, integ, origin, err string) {
	redirect(w, r, cid, integ, origin, "error", url.QueryEscape(err))
}

func extraData(t *oauth2.Token) map[string]any {
	extra := make(map[string]any)
	if v := t.Extra("instance_url"); v != nil {
		extra["instance_url"] = v // Salesforce
	}
	return extra
}
