package github

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

func (i integration) addIssueLabels(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		number      int
		labels      []string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"labels", &labels,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	c, _, err := gh.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdkvalues.Wrap(c)
}

func (i integration) removeIssueLabel(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, label string
		number             int
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"label", &label,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	c, err := gh.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdkvalues.Wrap(c)
}
