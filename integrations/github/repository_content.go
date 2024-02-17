package github

import (
	"context"

	"github.com/google/go-github/v54/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://docs.github.com/en/rest/repos/contents#create-or-update-file-contents
func (i integration) createOrUpdateFile(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, path, msg, branch, content, sha string
		committer                                    *[2]string
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"path", &path,
		"content", &content,
		"messsage", &msg,
		"sha?", &sha,
		"branch?", &branch,
		"committer?", &committer,
	)
	if err != nil {
		return nil, err
	}

	var opts github.RepositoryContentFileOptions

	if msg != "" {
		opts.Message = &msg
	}

	if content != "" {
		opts.Content = []byte(content)
	}

	if branch != "" {
		opts.Branch = &branch
	}

	if committer != nil {
		opts.Committer = &github.CommitAuthor{
			Name:  &committer[0],
			Email: &committer[1],
		}
	}

	if sha != "" {
		opts.SHA = &sha
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	c, _, err := gh.Repositories.CreateFile(ctx, owner, repo, path, &opts)
	if err != nil {
		return nil, err
	}

	return sdkvalues.Wrap(c)
}

// https://docs.github.com/en/rest/repos/contents#get-repository-content
func (i integration) getContents(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var owner, repo, path, ref string

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"path", &path,
		"ref?", &ref,
	)
	if err != nil {
		return nil, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	fileContent, directoryContent, _, err := gh.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, err
	}

	if directoryContent == nil && fileContent != nil {
		directoryContent = []*github.RepositoryContent{fileContent}
	}

	return sdkvalues.Wrap(directoryContent)
}
