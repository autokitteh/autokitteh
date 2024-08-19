package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/issues/issues#create-an-issue
func (i integration) createIssue(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		issue       github.IssueRequest
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"title", &issue.Title,
		"body?", &issue.Body,
		"assignee?", &issue.Assignee,
		"labels?", &issue.Labels,
		"assignees?", &issue.Assignees,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Issues.Create(ctx, owner, repo, &issue)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

// https://docs.github.com/en/rest/issues/issues#get-an-issue
func (i integration) getIssue(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		number      int
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

// https://docs.github.com/en/rest/issues/issues#update-an-issue
func (i integration) updateIssue(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		number      int

		issue github.IssueRequest
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"title?", &issue.Title,
		"body?", &issue.Body,
		"assignee?", &issue.Assignee,
		"state?", &issue.State,
		"state_reason?", &issue.StateReason,
		"milestone?", &issue.Milestone,
		"labels?", &issue.Labels,
		"assignees?", &issue.Assignees,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Issues.Edit(ctx, owner, repo, number, &issue)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

// https://docs.github.com/en/rest/issues/issues#list-repository-issues
func (i integration) listRepositoryIssues(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// TODO: Pagination.
	var (
		owner, repo string
		opts        github.IssueListByRepoOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"milestone?", &opts.Milestone,
		"state?", &opts.State,
		"assignee?", &opts.Assignee,
		"creator?", &opts.Creator,
		"mentioned?", &opts.Mentioned,
		"labels?", &opts.Labels,
		"sort?", &opts.Sort,
		"direction?", &opts.Direction,
		"since?", &opts.Since,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	is, _, err := gh.Issues.ListByRepo(ctx, owner, repo, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return valueWrapper.Wrap(is)
}
