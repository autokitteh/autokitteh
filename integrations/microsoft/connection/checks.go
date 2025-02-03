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
		case integrations.DaemonApp:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using daemon app"), nil
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
			if _, err = GetUserInfo(ctx, oauthToken(ctx, vs, o)); err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		case integrations.DaemonApp:
			if _, err = DaemonToken(ctx, vs); err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

// GetUserInfo returns the authenticated user's profile from Microsoft
// Graph (based on: https://learn.microsoft.com/en-us/graph/api/user-get).
func GetUserInfo(ctx context.Context, t *oauth2.Token) (*UserInfo, error) {
	u := "https://graph.microsoft.com/v1.0/me"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request for user info failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for user info failed: %s (%s)", resp.Status, body)
	}

	var user UserInfo
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &user, nil
}

type orgInfoWrapper struct {
	Value []OrgInfo `json:"value"`
}

// GetOrgInfo returns the authenticated user's organization profile from Microsoft
// Graph (based on: https://learn.microsoft.com/en-us/graph/api/organization-get).
func GetOrgInfo(ctx context.Context, t *oauth2.Token) (*OrgInfo, error) {
	u := "https://graph.microsoft.com/v1.0/organization"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request for org info failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read org info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for org info failed: %s (%s)", resp.Status, body)
	}

	var org orgInfoWrapper
	if err := json.Unmarshal(body, &org); err != nil {
		return nil, fmt.Errorf("failed to parse org info: %w", err)
	}
	if len(org.Value) != 1 {
		return nil, fmt.Errorf("unexpected number of Entra organizations: %s", body)
	}

	return &org.Value[0], nil
}
