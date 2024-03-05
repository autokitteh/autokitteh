package github

import (
	"context"

	"github.com/google/go-github/v54/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://docs.github.com/en/rest/commits/commits#list-commits
func (i integration) listCommits(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		opts        github.CommitsListOptions
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"opts?", &opts,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Repositories.ListCommits(ctx, owner, repo, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdkvalues.Wrap(resp)
}
