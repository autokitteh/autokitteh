package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/issues/comments#create-an-issue-comment
func (i integration) createIssueComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, body string
		number            int
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"number", &number,
		"body", &body,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	comment := &github.IssueComment{
		Body: github.String(body),
	}

	c, _, err := gh.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(c)
}
