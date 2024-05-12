package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/users#get-a-user
func (i integration) getUser(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var username, owner string

	err := sdkmodule.UnpackArgs(args, kwargs,
		"username", &username,
		"owner=?", &owner,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	var gh *github.Client

	if owner == "" {
		// According to github docs, this endpoint requires no permissions.
		gh, err = i.newAnonymousClient()
	} else {
		gh, err = i.NewClient(ctx, owner)
	}

	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Users.Get(ctx, username)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(resp)
}
