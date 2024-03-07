package github

import (
	"context"

	"github.com/google/go-github/v54/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/collaborators/collaborators#list-repository-collaborators
func (i integration) listCollaborators(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		opts        github.ListCollaboratorsOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"affiliation=?", &opts.Affiliation,
		"permission=?", &opts.Permission,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClientWithInstallJWT(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	c, _, err := gh.Repositories.ListCollaborators(ctx, owner, repo, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(c)
}
