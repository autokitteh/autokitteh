package zoom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// status checks the connection's initialization status (is it
// initialized? what type of authentication is configured?). This
// ensures that the connection is at least theoretically usable.
func status(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			return common.CheckOAuthToken(vs)
		case integrations.ServerToServer:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using S2S"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func test(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			return oauthTest(ctx, vs)
		case integrations.ServerToServer:
			return ServerToServerTest(ctx, vs)
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
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

// oauthTest verifies the OAuth authentication with Zoom API using the "me" context.
// (Based on: https://developers.zoom.us/docs/integrations/oauth/#the-me-context).
func oauthTest(ctx context.Context, vs sdktypes.Vars) (sdktypes.Status, error) {
	token := vs.GetValue(common.OAuthAccessTokenVar)
	refreshToken := vs.GetValue(common.OAuthRefreshTokenVar)
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")
	url := "https://api.zoom.us/v2/users/me"

	// First attempt with current token
	resp, err := makeAPICall(ctx, token, url)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OAuth connection successful"), nil
	}

	// Check if token expired
	if resp.StatusCode == http.StatusUnauthorized && refreshToken != "" {
		newToken, err := refreshTokenReq(ctx, refreshToken, clientID, clientSecret, vs)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		newResp, err := makeAPICall(ctx, newToken, url)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}
		defer newResp.Body.Close()

		if newResp.StatusCode == http.StatusOK {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OAuth connection successful after token refresh"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeError, "Failed to connect to Zoom API after token refresh"), nil
	}

	return sdktypes.NewStatus(sdktypes.StatusCodeError, "Failed to connect to Zoom API"), nil
}
