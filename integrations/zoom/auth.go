package zoom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

// serverToken retrieves a Server-to-Server (2-legged OAuth) token, using the connection's
// internal app details (based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func serverToken(ctx context.Context, app privateApp) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", app.AccountID)
	url := "https://zoom.us/oauth/token"

	resp, err := httpPost(ctx, data, app.ClientID, app.ClientSecret, url)
	if err != nil {
		return "", err
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
