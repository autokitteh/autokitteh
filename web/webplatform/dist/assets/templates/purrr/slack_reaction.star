"""Handler for Slack reaction events."""

load("@slack", "slack")
load("redis_helpers.star", "translate_slack_channel_id_to_pr_details")
load("user_helpers.star", "resolve_slack_user")

def on_slack_reaction_added(data):
    """https://api.slack.com/events/reaction_added

    Args:
        data: Slack event data.
    """
    owner, _, _ = translate_slack_channel_id_to_pr_details(data.item.channel)
    if not owner:
        return  # This is not a PR channel.
    github_user = resolve_slack_user(data.user, owner)
    msg = ":point_up: TODO - add GitHub review comment: `%s` added reaction `%s` (channel = `%s`, TS = `%s`)"
    msg %= (github_user, data.reaction, data.item.channel, data.item.ts)
    slack.chat_post_message(data.channel, msg)

    # Use GitHub's reactions API instead of comments? Easy, but requires user impersonation.
    # Reminder: in the GH reactions API, Slack "smile" = "laugh", Slack "tada" = "hurray".
    # Other supported reactions in GH: "+1", "-1", "confused", "heart", "rocket", "eyes".
    # See: https://docs.github.com/en/rest/reactions/reactions?apiVersion=2022-11-28
