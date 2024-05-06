package github

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/users#get-a-user
func (i integration) getUser(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var username string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"username", &username,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// According to github docs, this endpoint requires no permissions.
	gh, err := i.newAnonymousClient()
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Users.Get(ctx, username)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(resp)
}
