package github

import (
	"context"

	"github.com/google/go-github/v54/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://docs.github.com/en/rest/git/refs#create-a-reference
func (i integration) createRef(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var owner, repo, ref, sha string

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"ref", &ref,
		"sha", &sha,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	ghRef := github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: &sha,
		},
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	c, _, err := gh.Git.CreateRef(ctx, owner, repo, &ghRef)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdkvalues.Wrap(c)
}

// https://docs.github.com/en/rest/git/refs#get-a-reference
func (i integration) getRef(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var owner, repo, ref string

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"ref", &ref,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	c, _, err := gh.Git.GetRef(ctx, owner, repo, ref)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdkvalues.Wrap(c)
}
