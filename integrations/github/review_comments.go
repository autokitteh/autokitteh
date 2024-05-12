package github

import (
	"context"
	"time"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/pulls/comments#create-a-review-comment-for-a-pull-request
func (i integration) createReviewComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo          string
		pullNumber           int
		body, commitID, path *string

		line, startLine, inReplyTo   *int
		side, startSide, subjectType *string
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"body", &body,
		"commit_id", &commitID,
		"path", &path,

		"side?", &side,
		"line?", &line,
		"start_line?", &startLine,
		"start_side?", &startSide,
		"in_reply_to?", &inReplyTo,
		"subject_type?", &subjectType,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	var inReplyTo64 *int64
	if inReplyTo != nil {
		inReplyTo64 = new(int64)
		*inReplyTo64 = int64(*inReplyTo)
	}

	req := github.PullRequestComment{
		Body:     body,
		CommitID: commitID,
		Path:     path,

		Side:        side,
		Line:        line,
		StartLine:   startLine,
		StartSide:   startSide,
		InReplyTo:   inReplyTo64,
		SubjectType: subjectType,
	}

	comment, _, err := gh.PullRequests.CreateComment(ctx, owner, repo, pullNumber, &req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(comment)
}

// https://docs.github.com/en/rest/pulls/comments#create-a-reply-for-a-review-comment
func (i integration) createReviewCommentReply(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo, body     string
		pullNumber, commentID int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"comment_id", &commentID,
		"body", &body,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	comment, _, err := gh.PullRequests.CreateCommentInReplyTo(ctx, owner, repo, pullNumber, body, int64(commentID))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(comment)
}

// https://docs.github.com/en/rest/pulls/comments#delete-a-review-comment-for-a-pull-request
func (i integration) deleteReviewComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		commentID   int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"comment_id", &commentID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	_, err = gh.PullRequests.DeleteComment(ctx, owner, repo, int64(commentID))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(nil)
}

// https://docs.github.com/en/rest/pulls/comments#get-a-review-comment-for-a-pull-request
func (i integration) getReviewComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		commentID   int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"comment_id", &commentID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	comment, _, err := gh.PullRequests.GetComment(ctx, owner, repo, int64(commentID))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(comment)
}

// https://docs.github.com/en/rest/pulls/comments#list-review-comments-on-a-pull-request
func (i integration) listPullRequestReviewComments(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int

		sort, direction, since string
		perPage, page          int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,

		"sort?", &sort,
		"direction?", &direction,
		"since?", &since,
		"per_page?", &perPage,
		"page?", &page,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	opts := &github.PullRequestListCommentsOptions{}
	if sort != "" {
		opts.Sort = sort
	}
	if direction != "" {
		opts.Direction = direction
	}
	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err == nil {
			return sdktypes.InvalidValue, err
		}
		opts.Since = t
	}
	if perPage > 0 {
		opts.PerPage = perPage
	}
	if page > 0 {
		opts.Page = page
	}

	comments, _, err := gh.PullRequests.ListComments(ctx, owner, repo, pullNumber, opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(comments)
}

// https://docs.github.com/en/rest/pulls/comments#update-a-review-comment-for-a-pull-request
func (i integration) updateReviewComment(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		commentID   int
		body        *string
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"comment_id", &commentID,
		"body", &body,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	req := &github.PullRequestComment{
		Body: body,
	}

	comment, _, err := gh.PullRequests.EditComment(ctx, owner, repo, int64(commentID), req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(comment)
}
