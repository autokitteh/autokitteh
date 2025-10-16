package linear

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
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
		case integrations.APIKey:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using API key"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func test(v sdkservices.Vars, o *oauth.OAuth) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		body := `{ "query": "{ viewer { id  } }" }`
		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil

		case integrations.OAuthDefault, integrations.OAuthPrivate:
			token := o.FreshToken(ctx, zap.L(), desc, vs)
			_, err := common.HTTPPostJSON(ctx, linearAPIURL, token.AccessToken, body)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}

		case integrations.APIKey:
			apiKey := vs.GetValue(apiKeyVar)
			_, err := common.HTTPPostJSON(ctx, linearAPIURL, apiKey, body)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), err
			}

		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

// orgAndViewerInfo queries the Linear GraphQL API for important connection's details.
// A "viewer" is the authenticated user, and an "organization" is the user's workspace
// (based on: https://developers.linear.app/docs/graphql/working-with-the-graphql-api
// and: https://studio.apollographql.com/public/Linear-API/variant/current/home).
func orgAndViewerInfo(ctx context.Context, auth string) (*orgInfo, *viewerInfo, error) {
	url := "https://api.linear.app/graphql"
	query := "{ organization { id name urlKey } viewer { id displayName email name } }"
	resp, err := common.HTTPPostJSON(ctx, url, auth, fmt.Sprintf(`{"query": "%s"}`, query))
	if err != nil {
		return nil, nil, err
	}

	info := new(struct {
		Data struct {
			Org    orgInfo    `json:"organization"`
			Viewer viewerInfo `json:"viewer"`
		} `json:"data"`
	})
	if err := json.Unmarshal(resp, info); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &info.Data.Org, &info.Data.Viewer, nil
}
