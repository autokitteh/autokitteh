package loaders

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/go-github/v42/github"

	"github.com/autokitteh/L"
	"go.autokitteh.dev/sdk/api/apiprogram"
)

var GithubPathRewriter = NewPrefixPathRewriter("github", "github.com/", "github")

func NewGithubLoader(l L.L, getClient func(context.Context, string, string) (*github.Client, error)) LoaderFunc {
	return func(ctx context.Context, in *apiprogram.Path) ([]byte, string, error) {
		// github:org/repo/path...#ref

		parts := strings.SplitN(in.Path(), "/", 3)

		if len(parts) != 3 {
			return nil, "", fmt.Errorf("invalid path")
		}

		owner, repo, path := parts[0], parts[1], parts[2]

		l := L.N(l).With("owner", owner, "repo", repo, "path", path, "version", in.Version())

		l.Debug("downloading")

		// TODO: cache?
		var client *github.Client
		if getClient != nil {
			var err error
			if client, err = getClient(ctx, owner, repo); err != nil {
				return nil, "", fmt.Errorf("get github client: %w", err)
			}

			if client != nil {
				l.Debug("acquired installation specific client")
			}
		}

		if client == nil {
			client = github.NewClient(nil)
		}

		before := time.Now()

		stream, content, _, err := client.Repositories.DownloadContentsWithMeta(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: in.Version()})
		if err != nil {
			return nil, "", fmt.Errorf("download content: %w", err)
		}

		l = l.With("t", time.Since(before))

		defer stream.Close()

		bs, err := ioutil.ReadAll(stream)
		if err != nil {
			return nil, "", fmt.Errorf("read content: %w", err)
		}

		l.Debug("repo content downloaded", "size", len(bs))

		var sha string
		if content.SHA != nil {
			sha = *content.SHA
		}

		return bs, sha, nil
	}
}
