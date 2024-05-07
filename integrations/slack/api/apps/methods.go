package apps

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type API struct {
	Vars sdkservices.Vars
}

// ConnectionsOpenWithToken is only used internally, to check
// the usability of an app-level token provided by the user.
func ConnectionsOpenWithToken(ctx context.Context, vars sdkservices.Vars, appToken string) (*ConnectionsOpenResponse, error) {
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey, appToken)
	resp := &ConnectionsOpenResponse{}
	err := api.PostJSON(ctx, vars, struct{}{}, resp, "apps.connections.open")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}
