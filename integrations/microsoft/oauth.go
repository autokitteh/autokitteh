package microsoft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// oauthData contains OAuth token details.
type oauthData struct {
	AccessToken  string `var:"access_token,secret"`
	Expiry       string `var:"expiry"`
	RefreshToken string `var:"refresh_token,secret"`
	TokenType    string `var:"token_type"`
}

// userInfo contains user profile details from Microsoft Graph
// (based on: https://learn.microsoft.com/en-us/graph/api/user-get).
type userInfo struct {
	PrincipalName string `json:"userPrincipalName" var:"principal_name"`
	ID            string `json:"id" var:"id"`
	DisplayName   string `json:"displayName" var:"display_name"`
	Surname       string `json:"surname" var:"surname"`
	GivenName     string `json:"givenName" var:"given_name"`
	Language      string `json:"preferredLanguage" var:"language"`
	Mail          string `json:"mail" var:"mail"`
	MobilePhone   string `json:"mobilePhone" var:"mobile_phone"`
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
	vsid := sdktypes.NewVarScopeID(cid)

	// Handle OAuth errors (e.g. the user didn't authorize us), based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirection reported an error", zap.Error(errors.New(e)))
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

	// Test the OAuth token's usability by getting the user info associated with it.
	ctx := r.Context()
	user, err := h.getUserInfo(ctx, data.Token)
	if err != nil {
		l.Error("failed to fetch authenticated user details", zap.Error(err))
		c.AbortServerError("user details error")
		return
	}

	if err := h.saveConnection(ctx, vsid, data.Token, user); err != nil {
		l.Error("failed to save connection details", zap.Error(err))
		c.AbortServerError("failed to save connection details")
		return
	}

	http.Redirect(w, r, c.FinalURL(), http.StatusFound)
}

// getUserInfo returns user profile details from Microsoft Graph
// (based on: https://learn.microsoft.com/en-us/graph/api/user-get).
func (h handler) getUserInfo(ctx context.Context, t *oauth2.Token) (*userInfo, error) {
	url := "https://graph.microsoft.com/v1.0/me"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request for Microsoft user info failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for Microsoft user info failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Microsoft user info response: %w", err)
	}

	var user userInfo
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse Microsoft user info: %w", err)
	}

	return &user, nil
}

// saveConnection saves OAuth token and user profile details as connection variables.
func (h handler) saveConnection(ctx context.Context, vsid sdktypes.VarScopeID, t *oauth2.Token, u *userInfo) error {
	if t == nil {
		return errors.New("OAuth redirection missing token data")
	}

	vs := sdktypes.EncodeVars(oauthData{
		AccessToken:  t.AccessToken,
		Expiry:       t.Expiry.Format(time.RFC3339),
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
	}).WithPrefix("oauth_")

	vs = vs.Append(sdktypes.EncodeVars(u).WithPrefix("user_")...)

	return h.vars.Set(ctx, vs.WithScopeID(vsid)...)
}
