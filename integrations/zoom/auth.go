package zoom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	accountID    = sdktypes.NewSymbol("private_account_id")
	clientID     = sdktypes.NewSymbol("private_client_id")
	clientSecret = sdktypes.NewSymbol("private_client_secret")
)

// serverToken retrieves a Server-to-Server (2-legged OAuth) token, using the connection's
// internal app details (based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func serverToken(ctx context.Context, app *privateApp) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", app.AccountID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://zoom.us/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %v", err)
	}

	req.SetBasicAuth(app.ClientID, app.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: status %d: %s", resp.StatusCode, string(body))
	}

	var tokenData tokenResp

	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	return tokenData.AccessToken, nil
}

func makeAPICall(ctx context.Context, token string, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	return resp, nil
}

// refreshTokenReq sends a request to Zoom's OAuth endpoint to refresh an expired access token.
func refreshTokenReq(ctx context.Context, refreshToken string, clientID string, clientSecret string, vs sdktypes.Vars) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://zoom.us/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request: %v", err)
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	if err := vs.Set(common.OAuthAccessTokenVar, tokenResp.AccessToken, true); err != nil {
		return "", errors.New("Failed to save new access token")
	}
	if err := vs.Set(common.OAuthRefreshTokenVar, tokenResp.RefreshToken, true); err != nil {
		return "", errors.New("Failed to save new refresh token")
	}

	return tokenResp.AccessToken, nil
}
