"""Handler for Slack message events."""

load("@slack", "slack")
load("github_helpers.star", "create_review_comment", "create_review_comment_reply")
load("markdown.star", "slack_markdown_to_github")
load("redis_helpers.star", "translate_slack_channel_id_to_pr_details")
load("slack_helpers.star", "get_permalink")
load("user_helpers.star", "resolve_slack_user")

def on_slack_message(data):
    """https://api.slack.com/events/message

    Args:
        data: Slack event data.
    """
    subtype_handlers = {
        "": _on_slack_new_message,
        "message_changed": _on_slack_message_changed,
        "message_deleted": _on_slack_message_deleted,
        "thread_broadcast": _on_slack_new_message,
    }
    if data.subtype in subtype_handlers:
        subtype_handlers[data.subtype](data)

def _on_slack_new_message(data):
    """https://api.slack.com/events/message

    Args:
        data: Slack event data.
    """
    owner, repo, pr = translate_slack_channel_id_to_pr_details(data.channel)
    if not owner:
        return  # This is not a PR channel.
    github_user = resolve_slack_user(data.user, owner)

    # See subtype bug note in https://api.slack.com/events/message/message_replied
    if not data.thread_ts:
        # Slack message = GitHub review + single comment (we only need the comment
        # for correct 2-way syncs, but we can't have it without a parent review).
        review = "%s via %s" % (github_user, get_permalink(data.channel, data.ts))
        comment = slack_markdown_to_github(data.text, owner)
        create_review_comment(owner, repo, pr, review, comment, data.channel, data.ts)
    else:
        # Slack threaded reply = GitHub review comment.
        body = "%s replied via %s:\n\n"
        if not data.root:
            body %= (github_user, get_permalink(data.channel, data.ts))
        else:
            # Special case but same result: reply is broadcasted to the channel.
            body %= (github_user, get_permalink(data.channel, data.root.ts))
        body += slack_markdown_to_github(data.text, owner)

        create_review_comment_reply(owner, repo, pr, body, data.channel, data.thread_ts)

def _on_slack_message_changed(data):
    """https://api.slack.com/events/message/message_changed

    Args:
        data: Slack event data.
    """

    # Corner case 1: this event is also fired for a message when
    # a threaded reply is broadcasted to the channel, and when a
    # threaded reply is deleted - we don't care about them.
    if data.message.text == data.previous_message.text:
        return

    # Corner case 2: this event is also fired when a message
    # is deleted but its threaded replies are not - we handle
    # this as a regular deletion of a GitHub review.
    if data.message.subtype == "tombstone":
        msg = ":point_up: TODO - delete GitHub review (TS = `%s`)" % data.message.ts
        slack.chat_post_message(data.channel, msg)
        return

    github_user = resolve_slack_user(data.message.user)

    if not data.message.thread_ts:
        # Slack message = GitHub review.
        msg = ":point_up: TODO - edit GitHub review: `%s`: %s (TS = `%s`)"
        msg %= (github_user, data.message.text, data.message.ts)
    else:
        # Slack threaded reply = GitHub review comment.
        msg = ":point_up: TODO - edit GitHub review comment: `%s`: %s (TS = `%s`, Slack thread = `%s`)"
        msg %= (github_user, data.message.text, data.message.ts, data.message.thread_ts)

    slack.chat_post_message(data.channel, msg)

def _on_slack_message_deleted(data):
    """https://api.slack.com/events/message/message_deleted

    Args:
        data: Slack event data.
    """
    if not data.previous_message.thread_ts:
        # Slack message = GitHub review.
        msg = ":point_up: TODO - delete GitHub review (TS = `%s`)"
        msg %= data.deleted_ts
    else:
        # Slack threaded reply = GitHub review comment.
        msg = ":point_up: TODO - delete GitHub review comment (TS = `%s`, Slack thread = `%s`)"
        msg %= (data.deleted_ts, data.previous_message.thread_ts)  ###

    slack.chat_post_message(data.channel, msg)
