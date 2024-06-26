package auth

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type API struct {
	Vars sdkservices.Vars
}

// Test checks the caller's authentication & identity.
// Based on: https://api.slack.com/methods/auth.test.
// Required Slack app scopes: none.
func (a API) Test(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Invoke the API method.
	resp := &TestResponse{}
	err := api.PostJSON(ctx, a.Vars, struct{}{}, resp, "auth.test")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// TestWithToken is only used internally, when it's used by the OAuth redirect handler.
func TestWithToken(ctx context.Context, oauthToken string) (*TestResponse, error) {
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey, oauthToken)
	resp := &TestResponse{}
	err := api.PostJSON(ctx, nil, struct{}{}, resp, "auth.test")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}
