"""Handler for GitHub "issue_comment" events."""

load("debug.star", "debug")
load("markdown.star", "github_markdown_to_slack")
load(
    "redis_helpers.star",
    "map_github_link_to_slack_message_ts",
    "map_slack_message_ts_to_github_link",
)
load("slack_helpers.star", "impersonate_user_in_message", "lookup_pr_channel")

def on_github_issue_comment(data):
    """https://docs.github.com/webhooks/webhook-events-and-payloads#issue_comment

    This event occurs when there is activity relating
    to a comment on an issue or pull request.

    Args:
        data: GitHub event data.
    """

    # Ignore this event if it was triggered by a bot.
    if data.sender.type == "Bot":
        return

    action_handlers = {
        "created": _on_issue_comment_created,
        "edited": _on_issue_comment_edited,
        "deleted": _on_issue_comment_deleted,
    }
    if data.action in action_handlers:
        action_handlers[data.action](data)
    else:
        debug("Unrecognized GitHub issue comment action: `%s`" % data.action)

def _on_issue_comment_created(data):
    """A comment on an issue or pull request was created.

    Args:
        data: GitHub event data.
    """
    pr_url = data.issue.htmlurl
    org = data.organization.login
    channel_id = lookup_pr_channel(pr_url, data.issue.state)
    if not channel_id:
        debug("Can't sync this PR comment: " + data.comment.htmlurl)
        return

    msg = "<%s|PR comment>:\n\n" % data.comment.htmlurl
    msg += github_markdown_to_slack(data.comment.body, pr_url, org)
    thread_ts = impersonate_user_in_message(channel_id, data.sender, msg, org)
    if not thread_ts:
        return

    # Remember the thread timestamp (message ID) of the Slack message we posted.
    # Usage: syncing edits and deletes below to Slack.
    map_github_link_to_slack_message_ts(data.comment.htmlurl, thread_ts)

    # Also remember the GitHub comment URL, so we can reply to it later from Slack
    # (in create_review_comment_reply() in "github_helpers.star").
    channel_ts = "issue_comment:%s:%s" % (channel_id, thread_ts)
    map_slack_message_ts_to_github_link(channel_ts, data.comment.htmlurl)

def _on_issue_comment_edited(data):
    """A comment on an issue or pull request was edited.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.changes)
    print(data.issue)
    print(data.comment)
    print(data.sender)

def _on_issue_comment_deleted(data):
    """A comment on an issue or pull request was deleted.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.issue)
    print(data.comment)
    print(data.sender)
