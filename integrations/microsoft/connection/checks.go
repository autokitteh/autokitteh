// Package connection implements status and test checks
// that are reusable across all Microsoft integrations.
package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var AuthTypeVar = sdktypes.NewSymbol("auth_type")

// Status checks the connection's initialization status (is it
// initialized? what type of authentication is configured?). This
// ensures that the connection is at least theoretically usable.
func Status(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := v.Get(ctx, sdktypes.NewVarScopeID(cid), AuthTypeVar)
		if err != nil {
			return sdktypes.InvalidStatus, err // This is abnormal.
		}

		authType := vs.GetValue(AuthTypeVar)
		switch authType {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth 2.0"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// Test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func Test(v sdkservices.Vars, o sdkservices.OAuth) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := v.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err // This is abnormal.
		}

		authType := vs.GetValue(AuthTypeVar)
		switch authType {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			// Don't return, continue to do the check below.
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		// Load and attempt to use the OAuth token.
		if _, err = GetUserInfo(ctx, oauthToken(ctx, vs, o)); err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

// Return the OAuth token stored in the connection variables,
// unless it's stale, in which case it will be refreshed first.
func oauthToken(ctx context.Context, vs sdktypes.Vars, o sdkservices.OAuth) *oauth2.Token {
	t := &oauth2.Token{
		AccessToken:  vs.GetValueByString("oauth_access_token"),
		RefreshToken: vs.GetValueByString("oauth_refresh_token"),
		TokenType:    vs.GetValueByString("oauth_token_type"),
	}
	if t.Valid() {
		return t
	}

	cfg, _, err := o.Get(ctx, "microsoft")
	if err != nil {
		return t
	}

	t, err = cfg.TokenSource(ctx, t).Token()
	if err != nil {
		return nil
	}

	return t
}

// UserInfo contains user profile details from Microsoft Graph
// (based on: https://learn.microsoft.com/en-us/graph/api/user-get).
type UserInfo struct {
	PrincipalName string `json:"userPrincipalName" var:"principal_name"`
	ID            string `json:"id" var:"id"`
	DisplayName   string `json:"displayName" var:"display_name"`
	Surname       string `json:"surname" var:"surname"`
	GivenName     string `json:"givenName" var:"given_name"`
	Language      string `json:"preferredLanguage" var:"language"`
	Mail          string `json:"mail" var:"mail"`
	MobilePhone   string `json:"mobilePhone" var:"mobile_phone"`
}

// GetUserInfo returns the authenticated user's profile from Microsoft
// Graph (based on: https://learn.microsoft.com/en-us/graph/api/user-get).
func GetUserInfo(ctx context.Context, t *oauth2.Token) (*UserInfo, error) {
	url := "https://graph.microsoft.com/v1.0/me"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request for user info failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for user info failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	var user UserInfo
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &user, nil
}
