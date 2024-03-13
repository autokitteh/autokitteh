package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/actions/workflow-runs#list-workflow-runs-for-a-repository
func (i integration) listWorkflowRuns(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		opts        github.ListWorkflowRunsOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"branch=?", &opts.Branch,
		"event=?", &opts.Event,
		"actor=?", &opts.Actor,
		"status=?", &opts.Status,
		"created=?", &opts.Created,
		"head_sha=?", &opts.HeadSHA,
		"exclude_pull_requests=?", &opts.ExcludePullRequests,
		"check_suite_id=?", &opts.CheckSuiteID,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClientWithInstallJWT(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	c, _, err := gh.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(c)
}
