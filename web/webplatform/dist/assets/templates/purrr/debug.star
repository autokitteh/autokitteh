"""Simple, common utility for debugging and reporting errors."""

load("@slack", "slack")
load("env", "SLACK_DEBUG_CHANNEL")  # Set in "autokitteh.yaml".

def debug(msg):
    """Post a message to a special Slack channel, if defined.

    Args:
        msg: Message to post.
    """
    if not msg:
        return

    # Print the message in the autokitteh session's log.
    # This appears in the "ak session log" command's output.
    print(msg)

    if not SLACK_DEBUG_CHANNEL:
        return

    # This is more accessible than print().
    slack.chat_post_message(SLACK_DEBUG_CHANNEL, msg)
