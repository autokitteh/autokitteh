package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	vars sdkservices.Vars
}

var integrationID = sdktypes.NewIntegrationIDFromName("github")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "github",
	DisplayName:   "GitHub",
	Description:   "GitHub is a development platform with distributed version control, issue tracking, continuous integration, and more.",
	LogoUrl:       "/static/images/github.svg",
	ConnectionUrl: "/github/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(exportFuncs(i)...),
		connStatus(i),
		connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.OAuth:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using GitHub app"), nil
		case integrations.PAT:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using PAT + webhook"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

func connTest(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		var gh *github.Client
		if pat := vs.Get(vars.PAT); pat.IsValid() {
			gh = github.NewClient(nil).WithAuthToken(vs.GetValue(vars.PAT))
		} else {
			gh, err = newClientWithInstallJWT(vs)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		}

		// Make an API call that works for PAT & OAuth
		_, _, err = gh.RateLimit.Get(ctx)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

func exportFuncs(i *integration) []sdkmodule.Optfn {
	return []sdkmodule.Optfn{
		// Issues.
		sdkmodule.ExportFunction(
			"create_issue",
			i.createIssue,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/issues#create-an-issue"),
			sdkmodule.WithArgs("owner", "repo", "title", "body", "assignee", "milestone", "labels", "assignees"),
		),
		sdkmodule.ExportFunction(
			"get_issue",
			i.getIssue,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/issues#get-an-issue"),
			sdkmodule.WithArgs("owner", "repo", "number"),
		),
		sdkmodule.ExportFunction(
			"update_issue",
			i.updateIssue,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/issues#update-an-issue"),
			sdkmodule.WithArgs("owner", "repo", "number", "title", "body", "assignee", "state", "stateReason", "milestone", "labels", "assignees"),
		),
		sdkmodule.ExportFunction(
			"list_repository_issues",
			i.listRepositoryIssues,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/issues#list-repository-issues"),
			sdkmodule.WithArgs("owner", "repo", "milestone", "state", "assignee", "creator", "mentioned", "labels", "sort", "direction", "since"), // TODO: Pagination.
		),

		// Issue comments.
		sdkmodule.ExportFunction(
			"create_issue_comment",
			i.createIssueComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/comments#create-an-issue-comment"),
			sdkmodule.WithArgs("owner", "repo", "number", "body"),
		),
		// sdkmodule.ExportFunction("get_issue_comment",
		// 	sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/comments#get-an-issue-comment",
		// 	sdkmodule.WithArgs("owner", "repo", "number"),
		// 	i.getIssueComment),
		// sdkmodule.ExportFunction("update_issue_comment",
		// 	sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/comments#update-an-issue-comment",
		// 	sdkmodule.WithArgs("owner", "repo", "number"),
		// 	i.updateIssueComment),
		// sdkmodule.ExportFunction("list_issue_comments",
		// 	sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/comments#list-issue-comments",
		// 	sdkmodule.WithArgs("owner", "repo", "number"),
		// 	i.listIssueComments),

		// Issue labels.
		sdkmodule.ExportFunction(
			"add_issue_labels",
			i.addIssueLabels,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/labels#add-labels-to-an-issue"),
			sdkmodule.WithArgs("owner", "repo", "number", "labels"),
		),
		sdkmodule.ExportFunction(
			"remove_issue_label",
			i.removeIssueLabel,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/labels#remove-a-label-from-an-issue"),
			sdkmodule.WithArgs("owner", "repo", "number", "label"),
		),

		// sdkmodule.ExportFunction("set_issue_labels",
		// 	sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/labels#set-labels-for-an-issue",
		// 	sdkmodule.WithArgs("owner", "repo", "number"),
		// 	i.setIssueLabels),

		// sdkmodule.ExportFunction("remove_all_issue_labels",
		// 	sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/labels#remove-all-labels-from-an-issue",
		// 	sdkmodule.WithArgs("owner", "repo", "number"),
		// 	i.removeAllIssueLabels),
		// sdkmodule.ExportFunction("list_issue_labels",
		// 	sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/issues/labels#list-labels-for-an-issue",
		// 	sdkmodule.WithArgs("owner", "repo", "number"),
		// 	i.listIssueLabels),

		// Pull requests.
		sdkmodule.ExportFunction(
			"create_pull_request",
			i.createPullRequest,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "head", "base", "title?", "body?", "head_repo?", "draft?", "issue?", "maintainer_can_modify?"),
		),
		sdkmodule.ExportFunction(
			"get_pull_request",
			i.getPullRequest,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number"),
		),
		sdkmodule.ExportFunction(
			"list_pull_requests",
			i.listPullRequests,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/pulls#list-pull-requests"),
			sdkmodule.WithArgs("owner", "repo", "state", "head", "base", "sort", "direction", "per_page?", "page?"),
		),
		sdkmodule.ExportFunction(
			"list_pull_request_files",
			i.listPullRequestFiles,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/pulls#list-pull-requests-files"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "per_page?", "page?"),
		),
		sdkmodule.ExportFunction(
			"request_review",
			i.requestReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/review-requests#request-reviewers-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "reviewers=?", "team_reviewers=?"),
		),

		// Pull-request reviews.
		sdkmodule.ExportFunction(
			"create_review",
			i.createReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews?#create-a-review-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "commit_id?", "body?", "event?"),
		),
		sdkmodule.ExportFunction(
			"delete_pending_review",
			i.deletePendingReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#delete-a-pending-review-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "review_id"),
		),
		sdkmodule.ExportFunction(
			"dismiss_review",
			i.dismissReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#dismiss-a-review-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "review_id", "message"),
		),
		sdkmodule.ExportFunction(
			"get_review",
			i.getReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#get-a-review-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "review_id"),
		),
		sdkmodule.ExportFunction(
			"list_reviews",
			i.listReviews,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#list-reviews-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "per_page=?", "page=?"),
		),
		sdkmodule.ExportFunction(
			"list_review_comments",
			i.listReviewComments,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#list-comments-for-a-pull-request-review"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "review_id", "per_page=?", "page=?"),
		),
		sdkmodule.ExportFunction(
			"submit_review",
			i.submitReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#submit-a-review-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "review_id", "event", "body?"),
		),
		sdkmodule.ExportFunction(
			"update_review",
			i.updateReview,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/reviews#update-a-review-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "review_id", "body"),
		),

		// Pull-request review comments.
		sdkmodule.ExportFunction(
			"create_review_comment",
			i.createReviewComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/comments#create-a-review-comment-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "body", "commit_id", "path", "side?", "line?", "start_line?", "start_side?", "in_reply_to?", "subject_type?"),
		),
		sdkmodule.ExportFunction(
			"create_review_comment_reply",
			i.createReviewCommentReply,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/comments#create-a-reply-for-a-review-comment"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "comment_id", "body"),
		),
		sdkmodule.ExportFunction(
			"delete_review_comment",
			i.deleteReviewComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/comments#delete-a-review-comment-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "comment_id"),
		),
		sdkmodule.ExportFunction(
			"get_review_comment",
			i.getReviewComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/comments#get-a-review-comment-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "comment_id"),
		),
		sdkmodule.ExportFunction(
			"list_pr_review_comments",
			i.listPullRequestReviewComments,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/comments#list-review-comments-on-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "pull_number", "sort?", "direction?", "since?", "per_page?", "page?"),
		),
		sdkmodule.ExportFunction(
			"update_review_comment",
			i.updateReviewComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/pulls/comments#update-a-review-comment-for-a-pull-request"),
			sdkmodule.WithArgs("owner", "repo", "comment_id", "body"),
		),

		// Reactions.
		sdkmodule.ExportFunction(
			"create_reaction_for_commit_comment",
			i.createReactionForCommitComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-commit-comment"),
			sdkmodule.WithArgs("owner", "repo", "id", "content"),
		),
		sdkmodule.ExportFunction(
			"create_reaction_for_issue",
			i.createReactionForIssue,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue"),
			sdkmodule.WithArgs("owner", "repo", "number", "content"),
		),
		sdkmodule.ExportFunction(
			"create_reaction_for_issue_comment",
			i.createReactionForIssueComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue-comment"),
			sdkmodule.WithArgs("owner", "repo", "id", "content"),
		),
		sdkmodule.ExportFunction(
			"create_reaction_for_pull_request_review_comment",
			i.createReactionForPullRequestReviewComment,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-pull-request-review-comment"),
			sdkmodule.WithArgs("owner", "repo", "id", "content"),
		),

		// Repository Contents.
		sdkmodule.ExportFunction(
			"create_file",
			i.createOrUpdateFile,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/repos/contents#create-or-update-file-contents"),
			sdkmodule.WithArgs("owner", "repo", "path", "content", "message", "sha?", "branch?", "committer?"),
		),
		sdkmodule.ExportFunction(
			"get_contents",
			i.getContents,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/repos/contents#get-repository-content"),
			sdkmodule.WithArgs("owner", "repo", "path", "ref?"),
		),

		// Git references.
		sdkmodule.ExportFunction(
			"create_ref",
			i.createRef,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/git/refs#create-a-reference"),
			sdkmodule.WithArgs("owner", "repo", "ref", "sha"),
		),
		sdkmodule.ExportFunction(
			"get_ref",
			i.getRef,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/git/refs#get-a-reference"),
			sdkmodule.WithArgs("owner", "repo", "ref"),
		),

		// Repo
		sdkmodule.ExportFunction(
			"list_collaborators",
			i.listCollaborators,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/collaborators/collaborators#list-repository-collaborators"),
			sdkmodule.WithArgs(
				"owner", "repo", "affiliation=?", "permission=?",
			),
		),

		// Commits
		sdkmodule.ExportFunction(
			"list_commits",
			i.listCommits,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/commits/commits#list-commits"),
			sdkmodule.WithArgs("owner", "repo", "opts?"),
		),

		// Users
		sdkmodule.ExportFunction(
			"get_user",
			i.getUser,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/users#get-a-user"),
			sdkmodule.WithArgs("username", "owner=?"),
		),
		sdkmodule.ExportFunction(
			"search_users",
			i.searchUsers,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/search/search#search-users"),
			sdkmodule.WithArgs("query", "sort?", "order?", "per_page?", "page?", "owner=?"),
		),

		// Actions
		sdkmodule.ExportFunction(
			"list_workflows",
			i.listWorkflows,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/actions/workflows#list-repository-workflows"),
			sdkmodule.WithArgs("owner", "repo"),
		),
		sdkmodule.ExportFunction(
			"list_workflow_runs",
			i.listWorkflowRuns,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/actions/workflow-runs#list-workflow-runs-for-a-repository"),
			sdkmodule.WithArgs(
				"owner", "repo", "branch=?", "event=?", "actor=?", "status=?", "created=?",
				"head_sha=?", "exclude_pull_requests=?", "check_suite_id=?",
			),
		),
		sdkmodule.ExportFunction(
			"trigger_workflow",
			i.triggerWorkflow,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/actions/workflows#create-a-workflow-dispatch-event"),
			sdkmodule.WithArgs("owner", "repo", "ref", "workflow_file_name", "inputs?"),
		),

		// Checks
		sdkmodule.ExportFunction(
			"create_check_run",
			i.createCheckRun,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/checks/runs?create-a-check-run"),
			sdkmodule.WithArgs("owner", "repo", "name", "head_sha", "details_url?", "external_url?", "status?", "conclusion?", "output?", "created_at?", "completed_at?", "actions?"),
		),
		sdkmodule.ExportFunction(
			"update_check_run",
			i.updateCheckRun,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/checks/runs?update-a-check-run"),
			sdkmodule.WithArgs("owner", "repo", "check_run_id", "details_url?", "external_url?", "status?", "conclusion?", "output?", "created_at?", "completed_at?", "actions?"),
		),

		// Copilot
		sdkmodule.ExportFunction(
			"get_copilot_billing",
			i.getCopilotBilling,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#get-copilot-seat-information-and-settings-for-an-organization"),
			sdkmodule.WithArgs("org"),
		),
		sdkmodule.ExportFunction(
			"list_copilot_seats",
			i.listCopilotSeats,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#list-all-copilot-seat-assignments-for-an-organization"),
			sdkmodule.WithArgs("org", "page=?", "per_page=?"),
		),
		sdkmodule.ExportFunction(
			"add_copilot_teams",
			i.addCopilotTeams,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#add-teams-to-the-copilot-subscription-for-an-organization"),
			sdkmodule.WithArgs("org", "teams"),
		),
		sdkmodule.ExportFunction(
			"remove_copilot_teams",
			i.removeCopilotTeams,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#remove-teams-from-the-copilot-subscription-for-an-organization"),
			sdkmodule.WithArgs("org", "teams"),
		),
		sdkmodule.ExportFunction(
			"add_copilot_users",
			i.addCopilotUsers,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#add-users-to-the-copilot-subscription-for-an-organization"),
			sdkmodule.WithArgs("org", "users"),
		),
		sdkmodule.ExportFunction(
			"remove_copilot_users",
			i.removeCopilotUsers,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#remove-users-from-the-copilot-subscription-for-an-organization"),
			sdkmodule.WithArgs("org", "users"),
		),
		sdkmodule.ExportFunction(
			"get_copilot_seat_details",
			i.getCopilotSeatDetails,
			sdkmodule.WithFuncDoc("https://docs.github.com/en/rest/copilot/copilot-user-management#get-copilot-seat-assignment-details-for-a-user"),
			sdkmodule.WithArgs("org", "user"),
		),
	}
}
