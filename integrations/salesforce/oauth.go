package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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
		c.AbortBadRequest("request parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection after OAuth flow: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	// Handle OAuth errors (e.g. the user didn't authorize us), based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirection reported an error", zap.String("error", e))
		c.AbortBadRequest(e)
		return
	}

	// Decode the OAuth token.
	var data sdkintegrations.OAuthData
	err = kittehs.DecodeURLData(r.FormValue("oauth"), &data)
	if err != nil {
		l.Error("OAuth redirection returned invalid results", zap.Error(err))
		c.AbortServerError("invalid OAuth data")
		return
	}

	// Test the token's usability and get authoritative connection details.
	ctx := r.Context()
	accessToken := data.Token.AccessToken
	instanceURL := data.Extra["instance_url"].(string)
	userInfo, err := getUserInfo(ctx, instanceURL, accessToken)
	if err != nil {
		l.Error("failed to get user info", zap.Error(err))
		c.AbortServerError("failed to get user info")
		return
	}
	orgID := userInfo["organization_id"].(string)
	userID := userInfo["user_id"].(string)

	// Set the access token expiration time.
	clientID, err := h.accessTokenExpiration(ctx, instanceURL, data.Token, cid)
	if err != nil {
		l.Error("failed to get OAuth access token expiration", zap.Error(err))
		c.AbortServerError("failed to get OAuth access token expiration")
		return
	}

	vsid := sdktypes.NewVarScopeID(cid)
	if err := h.saveConnection(ctx, vsid, data.Token, data.Extra, orgID, userID); err != nil {
		l.Error("failed to save OAuth connection details", zap.Error(err))
		c.AbortServerError("failed to save connection details")
		return
	}

	h.subscribe(l, clientID, cid)

	// Redirect the user back to the UI.
	urlPath, err := c.FinalURL()
	if err != nil {
		l.Error("failed to construct final OAuth URL", zap.Error(err))
		c.AbortServerError("bad redirect URL")
		return
	}
	http.Redirect(w, r, urlPath, http.StatusFound)
}

// saveConnection saves OAuth token details as connection variables.
func (h handler) saveConnection(ctx context.Context, vsid sdktypes.VarScopeID, t *oauth2.Token, extra map[string]any, orgID, userID string) error {
	if t == nil {
		return errors.New("OAuth redirection missing token data")
	}

	vs := sdktypes.EncodeVars(common.EncodeOAuthData(t))
	for k, v := range extra {
		vs = vs.Append(sdktypes.NewVar(sdktypes.NewSymbol(k)).SetValue(fmt.Sprintf("%v", v)))
	}
	vs = vs.Append(sdktypes.NewVar(orgIDVar).SetValue(orgID))
	vs = vs.Append(sdktypes.NewVar(userIDVar).SetValue(userID))
	return h.vars.Set(ctx, vs.WithScopeID(vsid)...)
}

// accessTokenExpiration sets the expiration timestamp of the given access token.
// Salesforce access tokens are time-limited, but they don't have an expiry timestamp
// so we need to add it on our own. This function also returns the connection's client ID.
// Based on: https://help.salesforce.com/s/articleView?id=xcloud.remoteaccess_oidc_token_introspection_endpoint.htm
func (h handler) accessTokenExpiration(ctx context.Context, instanceURL string, t *oauth2.Token, cid sdktypes.ConnectionID) (string, error) {
	u, err := url.JoinPath(instanceURL, "/services/oauth2/introspect")
	if err != nil {
		return "", err
	}

	cfg, _, err := h.oauth.GetConfig(ctx, desc.UniqueName().String(), cid)
	if err != nil {
		return "", err
	}

	form := url.Values{
		"token":           {t.AccessToken},
		"token_type_hint": {"access_token"},
		"client_id":       {cfg.ClientID},
		"client_secret":   {cfg.ClientSecret},
	}

	resp, err := common.HTTPPostForm(ctx, u, "", form)
	if err != nil {
		return "", err
	}

	info := new(struct {
		Exp float64 `json:"exp"`
	})
	if err := json.Unmarshal(resp, info); err != nil {
		return "", err
	}

	t.Expiry = time.Unix(int64(info.Exp), 0)
	return cfg.ClientID, nil
}
