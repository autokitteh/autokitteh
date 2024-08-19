package github

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-commit-comment
func (i integration) createReactionForCommitComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, content string
		id                   int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"id", &id,
		"content", &content,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	r, _, err := gh.Reactions.CreateCommentReaction(ctx, owner, repo, id, content)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(r)
}

// https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue
func (i integration) createReactionForIssue(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, content string
		number               int
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"content", &content,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	r, _, err := gh.Reactions.CreateIssueReaction(ctx, owner, repo, number, content)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(r)
}

// https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue-comment
func (i integration) createReactionForIssueComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, content string
		id                   int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"id", &id,
		"content", &content,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	r, _, err := gh.Reactions.CreateIssueCommentReaction(ctx, owner, repo, id, content)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(r)
}

// https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-pull-request-review-comment
func (i integration) createReactionForPullRequestReviewComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, content string
		id                   int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"id", &id,
		"content", &content,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	r, _, err := gh.Reactions.CreatePullRequestCommentReaction(ctx, owner, repo, id, content)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(r)
}
