package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (i integration) createCheckRun(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		opts        github.CreateCheckRunOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		&opts,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Checks.CreateCheckRun(ctx, owner, repo, opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(resp)
}

func (i integration) updateCheckRun(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		checkRunID  int64
		opts        github.UpdateCheckRunOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"check_run_id", &checkRunID,
		&opts,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Checks.UpdateCheckRun(ctx, owner, repo, checkRunID, opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(resp)
}
