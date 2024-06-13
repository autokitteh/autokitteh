package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/pulls/reviews#create-a-review-for-a-pull-request
func (i integration) createReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int

		commitID, body, event *string
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,

		"commit_id?", &commitID,
		"body?", &body,
		"event?", &event,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	req := github.PullRequestReviewRequest{
		CommitID: commitID,
		Body:     body,
		Event:    event,
		// TODO: Comments: []*github.DraftReviewComment{},
	}

	review, _, err := gh.PullRequests.CreateReview(ctx, owner, repo, pullNumber, &req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(review)
}

// https://docs.github.com/en/rest/pulls/reviews#delete-a-pending-review-for-a-pull-request
func (i integration) deletePendingReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int
		reviewID    int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"review_id", &reviewID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	review, _, err := gh.PullRequests.DeletePendingReview(ctx, owner, repo, pullNumber, int64(reviewID))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(review)
}

// https://docs.github.com/en/rest/pulls/reviews#dismiss-a-review-for-a-pull-request
func (i integration) dismissReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int
		reviewID    int
		message     *string
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"review_id", &reviewID,
		"message", &message,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	req := github.PullRequestReviewDismissalRequest{
		Message: message,
	}

	review, _, err := gh.PullRequests.DismissReview(ctx, owner, repo, pullNumber, int64(reviewID), &req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(review)
}

// https://docs.github.com/en/rest/pulls/reviews#get-a-review-for-a-pull-request
func (i integration) getReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int
		reviewID    int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"review_id", &reviewID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	review, _, err := gh.PullRequests.GetReview(ctx, owner, repo, pullNumber, int64(reviewID))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(review)
}

// https://docs.github.com/en/rest/pulls/reviews#list-reviews-for-a-pull-request
func (i integration) listReviews(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int

		perPage, page int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,

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

	opts := &github.ListOptions{}
	if perPage > 0 {
		opts.PerPage = perPage
	}
	if page > 0 {
		opts.Page = page
	}

	reviews, _, err := gh.PullRequests.ListReviews(ctx, owner, repo, pullNumber, opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(reviews)
}

// https://docs.github.com/en/rest/pulls/reviews#list-comments-for-a-pull-request-review
func (i integration) listReviewComments(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int
		reviewID    int

		perPage, page int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"review_id", &reviewID,

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

	opts := &github.ListOptions{}
	if perPage > 0 {
		opts.PerPage = perPage
	}
	if page > 0 {
		opts.Page = page
	}

	comments, _, err := gh.PullRequests.ListReviewComments(ctx, owner, repo, pullNumber, int64(reviewID), opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(comments)
}

// https://docs.github.com/en/rest/pulls/reviews#submit-a-review-for-a-pull-request
func (i integration) submitReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo string
		pullNumber  int
		reviewID    int
		event       *string

		body *string
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"review_id", &reviewID,
		"event", &event,

		"body?", &body,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx, owner)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	req := github.PullRequestReviewRequest{
		Body:  body,
		Event: event,
	}

	review, _, err := gh.PullRequests.SubmitReview(ctx, owner, repo, pullNumber, int64(reviewID), &req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(review)
}

// https://docs.github.com/en/rest/pulls/reviews#update-a-review-for-a-pull-request
func (i integration) updateReview(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		owner, repo, body    string
		pullNumber, reviewID int
	)

	err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"pull_number", &pullNumber,
		"review_id", &reviewID,
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

	review, _, err := gh.PullRequests.UpdateReview(ctx, owner, repo, pullNumber, int64(reviewID), body)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(review)
}
