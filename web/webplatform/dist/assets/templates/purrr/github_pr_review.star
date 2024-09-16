"""Handler for GitHub "pull_request_review" events."""

load("debug.star", "debug")
load("markdown.star", "github_markdown_to_slack")
load(
    "redis_helpers.star",
    "map_github_link_to_slack_message_ts",
    "map_slack_message_ts_to_github_link",
)
load(
    "slack_helpers.star",
    "impersonate_user_in_message",
    "lookup_pr_channel",
    "mention_user_in_message",
)

def on_github_pull_request_review(data):
    """https://docs.github.com/webhooks/webhook-events-and-payloads#pull_request_review

    A pull request review is a group of pull request review
    comments in addition to a body comment and a state.

    For more information, see "About pull request reviews":
    https://docs.github.com/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews

    Args:
        data: GitHub event data.
    """

    # Ignore this event if it was triggered by a bot.
    if data.sender.type == "Bot":
        return

    action_handlers = {
        "submitted": _on_pr_review_submitted,
        "edited": _on_pr_review_edited,
        "dismissed": _on_pr_review_dismissed,
    }
    if data.action in action_handlers:
        action_handlers[data.action](data)
    else:
        debug("Unrecognized GitHub PR review action: `%s`" % data.action)

def _on_pr_review_submitted(data):
    """A review on a pull request was submitted.

    This is usually not interesting in itself, unless the review
    state is "approved", and/or the review body isn't empty.

    Args:
        data: GitHub event data.
    """
    org = data.organization.login
    pr_url = data.pull_request.htmlurl
    channel_id = lookup_pr_channel(pr_url, data.pull_request.state)
    if not channel_id:
        debug("Can't sync this PR review: " + data.review.htmlurl)
        return

    if data.review.state == "approved":
        if not data.review.body:
            msg = "%s approved this PR :+1:"
            mention_user_in_message(channel_id, data.sender, msg, org)
            return
        else:
            msg = "<%s|PR approved> :+1:\n\n" % data.review.htmlurl
    elif data.review.body:
        msg = "<%s|PR review>:\n\n" % data.review.htmlurl
    else:
        return

    msg += github_markdown_to_slack(data.review.body, pr_url, org)
    thread_ts = impersonate_user_in_message(channel_id, data.sender, msg, org)
    if not thread_ts:
        return

    # Remember the thread timestamp (message ID) of the Slack message we posted.
    # Usage: syncing edits below to Slack.
    map_github_link_to_slack_message_ts(data.review.htmlurl, thread_ts)

    # Also remember the GitHub review URL, so we can reply to it later from Slack
    # (in create_review_comment_reply() in "github_helpers.star").
    channel_ts = "review:%s:%s" % (channel_id, thread_ts)
    map_slack_message_ts_to_github_link(channel_ts, data.review.htmlurl)

def _on_pr_review_edited(data):
    """The body comment on a pull request review was edited.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    if not getattr(data, "changes", None):
        return

    print(data.changes)
    print(data.review)
    print(data.sender)
    print(data.pull_request)

def _on_pr_review_dismissed(data):
    """A review on a pull request was dismissed.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.review)
    print(data.sender)
    print(data.pull_request)
