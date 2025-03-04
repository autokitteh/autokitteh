package zoom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// serverToken retrieves a Server-to-Server (2-legged OAuth) token, using the connection's
// internal app details (based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func serverToken(ctx context.Context, app privateApp) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", app.AccountID)

	resp, err := httpPost(ctx, data, app.ClientID, app.ClientSecret)
	if err != nil {
		return "", fmt.Errorf("refresh token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: status %d: %s", resp.StatusCode, string(body))
	}

	var tokenData common.OAuthData
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	return tokenData.AccessToken, nil
}

// refreshToken sends a request to Zoom's OAuth endpoint to refresh an expired access token.
func refreshToken(ctx context.Context, refreshT string, clientID string, clientSecret string, vs sdktypes.Vars) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshT)

	resp, err := httpPost(ctx, data, clientID, clientSecret)
	if err != nil {
		return "", fmt.Errorf("refresh token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to refresh token")
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	// TODO: var tokenResp common.OAuthData
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	vs.Set(common.OAuthAccessTokenVar, tokenResp.AccessToken, true) // TODO: is this correct

	vs.Set(common.OAuthRefreshTokenVar, tokenResp.RefreshToken, true)
	// TODO: vs.Set(common.OAuthExpiryVar, tokenResp.Expiry, false);  add json:"expiry" to OAuthData struct

	return tokenResp.AccessToken, nil
}
