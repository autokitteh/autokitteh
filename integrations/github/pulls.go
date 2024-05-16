package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request
func (i integration) createPullRequest(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo, head, base string
		title, body, headRepo   *string
		draft, mcm              *bool
		issue                   *int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"head", &head,
		"base", &base,

		"title?", &title,
		"body?", &body,
		"head_repo?", &headRepo,
		"draft?", &draft,
		"issue?", &issue,
		"maintainer_can_modify?", &mcm,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	pull := github.NewPullRequest{
		Title:               title,
		Head:                &head,
		HeadRepo:            headRepo,
		Base:                &base,
		Body:                body,
		Issue:               issue,
		MaintainerCanModify: mcm,
		Draft:               draft,
	}

	pr, _, err := gh.PullRequests.Create(ctx, owner, repo, &pull)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(pr)
}

// https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request
func (i integration) getPullRequest(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, pullNumber)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(pr)
}

// https://docs.github.com/en/rest/pulls/pulls#list-pull-requests
func (i integration) listPullRequests(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string

		opts github.PullRequestListOptions
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,

		"state?", &opts.State,
		"head?", &opts.Head,
		"base?", &opts.Base,
		"sort?", &opts.Sort,
		"direction?", &opts.Direction,
		"per_page?", &opts.ListOptions.PerPage,
		"page?", &opts.ListOptions.Page,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	prs, _, err := gh.PullRequests.List(ctx, owner, repo, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(prs)
}

// https://docs.github.com/en/rest/pulls/pulls#list-pull-requests-files
func (i integration) listPullRequestFiles(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int

		opts github.ListOptions
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,

		"per_page?", &opts.PerPage,
		"page?", &opts.Page,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	files, _, err := gh.PullRequests.ListFiles(ctx, owner, repo, pullNumber, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(files)
}

// https://docs.github.com/en/rest/pulls/review-requests#request-reviewers-for-a-pull-request
func (i integration) requestReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int

		req github.ReviewersRequest
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,

		"reviewers=?", &req.Reviewers,
		"team_reviewers=?", &req.TeamReviewers,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	pr, _, err := gh.PullRequests.RequestReviewers(ctx, owner, repo, pullNumber, req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(pr)
}
