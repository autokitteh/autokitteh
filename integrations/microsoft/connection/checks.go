// Package connection implements status and test checks
// that are reusable across all Microsoft integrations.
package connection

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Status checks the connection's initialization status (is it
// initialized? what type of authentication is configured?). This
// ensures that the connection is at least theoretically usable.
func Status(v sdkservices.Vars) sdkintegrations.OptFn {
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
		case integrations.DaemonApp:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using daemon app"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// Test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func Test(v sdkservices.Vars, o *oauth.OAuth) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch authType := common.ReadAuthType(vs); authType {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil

		case integrations.OAuthDefault, integrations.OAuthPrivate:
			desc := common.Descriptor("microsoft", "", "")
			l := zap.L().With(
				zap.String("integration", desc.UniqueName().String()),
				zap.String("connection_id", cid.String()),
				zap.String("auth_type", authType),
			)
			t := o.FreshToken(ctx, l, desc, vs)
			if _, err = GetUserInfo(ctx, t); err != nil {
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
	resp, err := common.HTTPGet(ctx, u, "Bearer "+t.AccessToken)
	if err != nil {
		return nil, err
	}

	var user UserInfo
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &user, nil
}

// GetOrgInfo returns the authenticated user's organization profile from Microsoft
// Graph (based on: https://learn.microsoft.com/en-us/graph/api/organization-get).
func GetOrgInfo(ctx context.Context, t *oauth2.Token) (*OrgInfo, error) {
	u := "https://graph.microsoft.com/v1.0/organization"
	resp, err := common.HTTPGet(ctx, u, "Bearer "+t.AccessToken)
	if err != nil {
		return nil, err
	}

	org := new(struct {
		Value []OrgInfo `json:"value"`
	})
	if err := json.Unmarshal(resp, org); err != nil {
		return nil, fmt.Errorf("failed to parse org info: %w", err)
	}
	if len(org.Value) != 1 {
		return nil, fmt.Errorf("unexpected number of Entra organizations: %s", resp)
	}

	return &org.Value[0], nil
}
