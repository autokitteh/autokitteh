package github

import (
	"context"
	"time"

	"github.com/google/go-github/v54/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://docs.github.com/en/rest/pulls/comments#list-review-comments-on-a-pull-request
func (i integration) listPullRequestReviewComments(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// TODO: Pagination.
	var (
		owner, repo, sort, direction, since string
		number                              int
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"sort?", &sort,
		"direction?", &direction,
		"since?", &since,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	opts := &github.PullRequestListCommentsOptions{}

	if sort != "" {
		opts.Sort = sort
	}

	if direction != "" {
		opts.Direction = direction
	}

	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err == nil {
			return sdktypes.InvalidValue, err
		}
		opts.Since = t
	}

	cs, _, err := gh.PullRequests.ListComments(ctx, owner, repo, number, opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdkvalues.Wrap(cs)
}
