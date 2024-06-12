package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/users#get-a-user
func (i integration) getUser(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
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

	// Invoke the API method.
	resp, _, err := gh.Users.Get(ctx, username)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

// https://docs.github.com/en/rest/search/search#search-users
// https://docs.github.com/en/search-github/searching-on-github/searching-users
func (i integration) searchUsers(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		query, owner string
		opts         github.SearchOptions
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"query", &query,
		"sort?", &opts.Sort,
		"order?", &opts.Order,
		"per_page?", &opts.PerPage,
		"page?", &opts.Page,
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

	// Invoke the API method.
	resp, _, err := gh.Search.Users(ctx, query, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}
