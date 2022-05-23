package githubplugin

import (
	"context"
	"fmt"

	"github.com/google/go-github/v42/github"

	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"

	"github.com/autokitteh/autokitteh/internal/pkg/githubinstalls"
)

var (
	Plugin = &pluginimpl.Plugin{
		ID:  "github",
		Doc: "github plugin",
		Members: map[string]*pluginimpl.PluginMember{
			"open": pluginimpl.NewMethodMember(
				"TODO",
				func(
					ctx context.Context,
					name string,
					args []*apivalues.Value,
					kwargs map[string]*apivalues.Value,
					funcToValue pluginimpl.FuncToValueFunc,
				) (*apivalues.Value, error) {
					var repo, owner string

					if err := pluginimpl.UnpackArgs(
						args, kwargs,
						"owner", &owner,
						"repo", &repo,
					); err != nil {
						return nil, err
					}

					client, err := githubinstalls.GetInstalls().GetClient(ctx, owner, repo)
					if err != nil {
						return nil, fmt.Errorf("get install: %w", err)
					}

					if client == nil {
						client = github.NewClient(nil)
					}

					return pluginimpl.BuildStruct(
						funcToValue,
						"github.client",
						pluginimpl.NewStructSimpleFuncMember(
							"add_label_to_issue",
							"TODO",
							func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
								var (
									number int
									label  string
								)

								if err := pluginimpl.UnpackArgs(
									args, kwargs,
									"number", &number,
									"label", &label,
								); err != nil {
									return nil, err
								}

								labels, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, number, []string{label})

								if err != nil {
									return nil, err
								}

								return apivalues.Wrap(labels)
							},
						),
						pluginimpl.NewStructSimpleFuncMember(
							"remove_label_for_issue",
							"TODO",
							func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
								var (
									number int
									label  string
								)

								if err := pluginimpl.UnpackArgs(
									args, kwargs,
									"number", &number,
									"label", &label,
								); err != nil {
									return nil, err
								}

								_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)

								return apivalues.None, err
							},
						),
						pluginimpl.NewStructSimpleFuncMember(
							"list_labels_by_issue",
							"TODO",
							func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
								var number int

								if err := pluginimpl.UnpackArgs(
									args, kwargs,
									"number", &number,
								); err != nil {
									return nil, err
								}

								labels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, number, nil)
								if err != nil {
									return nil, err
								}

								return apivalues.Wrap(labels)
							},
						),
						pluginimpl.NewStructSimpleFuncMember(
							"issue_create_comment",
							"TODO",
							func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
								var (
									number int
									text   string
								)

								if err := pluginimpl.UnpackArgs(
									args, kwargs,
									"number", &number,
									"text", &text,
								); err != nil {
									return nil, err
								}

								comment, _, err := client.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
									Body: &text,
								})

								if err != nil {
									return nil, err
								}

								return apivalues.Wrap(comment)
							},
						),
					)
				},
			),
		},
	}
)
