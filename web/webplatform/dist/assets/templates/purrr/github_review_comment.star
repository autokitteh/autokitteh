"""Handler for GitHub "pull_request_review_comment" events."""

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
    "impersonate_user_in_reply",
    "lookup_pr_channel",
)

def on_github_pull_request_review_comment(data):
    """https://docs.github.com/webhooks/webhook-events-and-payloads#pull_request_review_comment

    A pull request review comment is a comment on a pull request's diff.

    For more information, see "Commenting on a pull request":
    https://docs.github.com/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/commenting-on-a-pull-request#adding-line-comments-to-a-pull-request

    Args:
        data: GitHub event data.
    """

    # Ignore this event if it was triggered by a bot.
    if data.sender.type == "Bot":
        return

    action_handlers = {
        "created": _on_pr_review_comment_created,
        "edited": _on_pr_review_comment_edited,
        "deleted": _on_pr_review_comment_deleted,
    }
    if data.action in action_handlers:
        action_handlers[data.action](data)
    else:
        debug("Unrecognized GitHub PR review comment action: `%s`" % data.action)

def _on_pr_review_comment_created(data):
    """A comment on a pull request diff was created.

    Args:
        data: GitHub event data.
    """
    org = data.org.login
    pr_url = data.pull_request.htmlurl
    channel_id = lookup_pr_channel(pr_url, data.pull_request.state)
    if not channel_id:
        debug("Can't sync this PR review comment: " + data.comment.htmlurl)
        return

    if not getattr(data.comment, "in_reply_to", None):
        # Review comment.
        msg = "<%s|%s review comment> in `%s`:\n\n"
        msg %= (data.comment.htmlurl, data.comment.subject_type.capitalize(), data.comment.path)
        msg += github_markdown_to_slack(data.comment.body, pr_url, org)
        thread_ts = impersonate_user_in_message(channel_id, data.sender, msg, org)

        # Remember the GitHub comment ID, so we can reply to it later from Slack.
        # See usage in create_review_comment_reply() in "github_helpers.star".
        if thread_ts:
            channel_ts = "review_comment:%s:%s" % (channel_id, thread_ts)
            map_slack_message_ts_to_github_link(channel_ts, data.comment.id)
    else:
        # Review comment in reply to another review comment.
        thread_url = "%s#discussion_r%d" % (pr_url, data.comment.in_reply_to)
        msg = "<%s|Reply to review comment>:\n\n" % data.comment.htmlurl
        msg += github_markdown_to_slack(data.comment.body, pr_url, org)
        thread_ts = impersonate_user_in_reply(channel_id, thread_url, data.sender, msg, org)

    # Remember the thread/reply timestamp (message ID) of the Slack message we posted.
    # Usage: edit and delete below, impersonate_user_in_reply() for syncing replies.
    if thread_ts:
        map_github_link_to_slack_message_ts(data.comment.htmlurl, thread_ts)

def _on_pr_review_comment_edited(data):
    """The content of a comment on a pull request diff was changed.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.changes)
    print(data.comment)
    print(data.sender)
    print(data.pull_request)

def _on_pr_review_comment_deleted(data):
    """A comment on a pull request diff was deleted.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.comment)
    print(data.sender)
    print(data.pull_request)
