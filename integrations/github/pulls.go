package github

import (
	"context"

	"github.com/google/go-github/v54/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request
func (i integration) getPullRequest(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		number      int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}

	return sdkvalues.Wrap(pr)
}

// https://docs.github.com/en/rest/pulls/pulls#list-pull-requests
func (i integration) listPullRequests(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// TODO: Pagination.
	var (
		owner, repo                        string
		head, base, sort, direction, state string
		page, perPage                      int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"state?", &state,
		"head?", &head,
		"base?", &base,
		"sort?", &sort,
		"direction?", &direction,
		"page?", &page,
		"per_page?", &perPage,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	opts := &github.PullRequestListOptions{
		State:     state,
		Head:      head,
		Base:      base,
		Sort:      sort,
		Direction: direction,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	prs, _, err := gh.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	return sdkvalues.Wrap(prs)
}

// https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request
func (i integration) createPullRequest(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo, head, base string
		body, title, headRepo   *string
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
		return nil, err
	}

	pull := github.NewPullRequest{
		Title:               title,
		Head:                &head,
		Base:                &base,
		HeadRepo:            headRepo,
		Body:                body,
		Issue:               issue,
		Draft:               draft,
		MaintainerCanModify: mcm,
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	pr, _, err := gh.PullRequests.Create(ctx, owner, repo, &pull)
	if err != nil {
		return nil, err
	}

	return sdkvalues.Wrap(pr)
}

// https://docs.github.com/en/rest/pulls/review-requests#request-reviewers-for-a-pull-request
func (i integration) requestReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		number      int
		request     github.ReviewersRequest
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"reviewers=?", &request.Reviewers,
		"team_reviewers=?", &request.TeamReviewers,
	)
	if err != nil {
		return nil, err
	}

	gh, err := i.NewClientWithInstallJWT(ctx)
	if err != nil {
		return nil, err
	}

	pr, _, err := gh.PullRequests.RequestReviewers(ctx, owner, repo, number, request)
	if err != nil {
		return nil, err
	}

	return sdkvalues.Wrap(pr)
}
