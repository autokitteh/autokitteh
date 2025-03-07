package linear

import (
	"context"
	"encoding/json"
	"fmt"

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
		vs, errStatus, err := common.ReadConnectionVars(ctx, v, cid)
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
func test(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadConnectionVars(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			// TODO(INT-269): Implement.
		case integrations.APIKey:
			// TODO(INT-269): Implement.
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		// TODO(INT-269): return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
		return sdktypes.NewStatus(sdktypes.StatusCodeError, "Not implemented"), nil
	})
}

// orgAndViewerInfo queries the Linear GraphQL API for important connection's details
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
