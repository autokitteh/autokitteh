package zoom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	acountID     = sdktypes.NewSymbol("private_account_id")
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

// ServerToServerTest handles the Server-to-Server authentication test for Zoom
// (Based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func ServerToServerTest(ctx context.Context, vs sdktypes.Vars) (sdktypes.Status, error) {
	id := vs.GetValue(acountID)
	clientID := vs.GetValue(clientID)
	secret := vs.GetValue(clientSecret)

	app := privateApp{
		AccountID:    id,
		ClientID:     clientID,
		ClientSecret: secret,
	}

	// Get a server-to-server token
	token, err := serverToken(ctx, &app)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}

	url := "https://api.zoom.us/v2/users"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, "API request failed"), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Server-to-Server connection successful"), nil
	}

	return sdktypes.NewStatus(sdktypes.StatusCodeError, "Failed to connect to Zoom API"), nil
}
