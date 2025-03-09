package salesforce

import (
	"context"
	"encoding/json"
	"net/url"

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
			// TODO(INT-268): Implement.
		case integrations.APIKey:
			// TODO(INT-268): Implement.
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		// TODO(INT-235): return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
		return sdktypes.NewStatus(sdktypes.StatusCodeError, "Not implemented"), nil
	})
}

// https://help.salesforce.com/s/articleView?id=sf.remoteaccess_using_userinfo_endpoint.htm
func getUserInfo(ctx context.Context, instanceURL, accessToken string) (map[string]any, error) {
	u, err := url.JoinPath(instanceURL, "services/oauth2/userinfo")
	if err != nil {
		return nil, err
	}

	resp, err := common.HTTPGet(ctx, u, "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]any
	if err := json.Unmarshal(resp, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
