package zoom

import (
	"context"
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

// oauthTest verifies the OAuth authentication with Zoom API using the "me" context.
// (Based on: https://developers.zoom.us/docs/integrations/oauth/#the-me-context).
func oauthTest(ctx context.Context, vs sdktypes.Vars) (sdktypes.Status, error) {
	data := common.OAuthData{}
	vs.Decode(&data)
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")
	url := "https://api.zoom.us/v2/users/me"

	// Check if the token is expired and refresh it.
	expiry, err := time.Parse(time.RFC3339, data.Expiry)
	if err == nil {
		if time.Now().After(expiry) && data.RefreshToken != "" {
			data.AccessToken, err = refreshToken(ctx, data.RefreshToken, clientID, clientSecret, vs)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		}
	}

	resp, err := httpGet(ctx, data.AccessToken, url)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OAuth connection successful"), nil
	}

	return sdktypes.NewStatus(sdktypes.StatusCodeError, "Failed to connect to Zoom API"), nil
}

// ServerToServerTest validates the Server-to-Server authentication test for Zoom.
// (Based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func ServerToServerTest(ctx context.Context, vs sdktypes.Vars) (sdktypes.Status, error) {
	var app privateApp
	vs.Decode(&app)

	// Get a server-to-server token.
	token, err := serverToken(ctx, app)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}

	url := "https://api.zoom.us/v2/users"
	resp, err := httpGet(ctx, token, url)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Server-to-Server connection successful"), nil
	}

	return sdktypes.NewStatus(sdktypes.StatusCodeError, "Failed to connect to Zoom API"), nil
}

// httpGet sends an HTTP GET request to the specified URL with a Bearer token for authentication.
func httpGet(ctx context.Context, bearerToken string, url string) (*http.Response, error) {
	timoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(timoutCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	return resp, nil
}

// httpPost sends an HTTP POST request with a URL-encoded body and basic authentication.
func httpPost(ctx context.Context, data url.Values, clientID, clientSecret string) (*http.Response, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	body := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, "https://zoom.us/oauth/token", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if clientID != "" && clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return resp, nil
}
