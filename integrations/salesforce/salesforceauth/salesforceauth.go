package salesforceauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// AccessTokenExpiration sends a request to the given instance's OAuth2 introspection endpoint to retrieve
// the expiration timestamp of the token and updates the provided token's expiry field.
// Returns an error if the request fails, the response is invalid, or the expiration timestamp is missing.
func AccessTokenExpiration(ctx context.Context, instanceURL string, t *oauth2.Token, vsid sdktypes.VarScopeID, vars sdkservices.Vars) error {
	vs, err := vars.Get(ctx, vsid)
	if err != nil {
		return err
	}

	formData := url.Values{
		"token":           {t.AccessToken},
		"token_type_hint": {"access_token"},
		"client_id":       {vs.GetValueByString("private_client_id")},
		"client_secret":   {vs.GetValueByString("private_client_secret")},
	}

	u, err := url.JoinPath(instanceURL, "services/oauth2/introspect")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return errors.New("failed to parse token info")
	}

	// Extract the expiration timestamp.
	expFloat, ok := tokenInfo["exp"].(float64)
	if !ok {
		return errors.New("missing or invalid expiration time in response")
	}
	t.Expiry = time.Unix(int64(expFloat), 0)

	return nil
}
