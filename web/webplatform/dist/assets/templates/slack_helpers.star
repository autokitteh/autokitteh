"""Slack API helper functions."""

load("@slack", "slack")
load("debug.star", "debug")
load("env", "SLACK_CHANNEL_PREFIX", "SLACK_LOG_CHANNEL")  # Set in "autokitteh.yaml".
load("redis_helpers.star", "get_slack_opt_out", "lookup_github_link_details")
load("user_helpers.star", "github_username_to_slack_user", "resolve_github_user")

_CHANNEL_MAX_METADATA_LENGTH = 250  # Characters.

def add_users_to_channel(channel_id, users):
    """Invite all the participants in a GitHub PR to a Slack channel.

    Args:
        channel_id: Slack channel ID.
        users: Comma-separated list of (up to 1000) Slack user IDs.
    """

    # Quietly ignore users who opted out of PuRRR. They will still be
    # mentioned in the channel, but as non-members they won't know it.
    opted_in = []
    for user_id in users.split(","):
        if not get_slack_opt_out(user_id):
            opted_in.append(user_id)
    users = ",".join(opted_in)

    # See: https://api.slack.com/methods/conversations.invite
    resp = slack.conversations_invite(channel_id, users, force = True)
    if resp.ok or resp.error == "already_in_channel":
        return

    # An error occurred - first, report it.
    debug("Add Slack users to channel <#%s>: `%s`" % (channel_id, resp.error))
    for e in resp.errors:
        debug("Error: <@%s> - `%s`" % (e.user, e.error))

    # Now check if it's fatal or not.
    # See: https://api.slack.com/methods/conversations.members
    resp = slack.conversations_members(channel_id, limit = 100)
    if resp.ok and len(resp.members) > 1:  # At least some users were added.
        for user_id in resp.members:
            debug("Member: <@%s>" % user_id)
        return

    # No members at all? Abort the channel.
    # See: https://api.slack.com/methods/conversations.archive
    resp = slack.conversations_archive(channel_id)
    if not resp.ok:
        debug("Archive DOA channel `%s`: `%s`" % (channel_id, resp.error))

def archive_channel(channel_id, data):
    """Archive a Slack channel.

    Args:
        channel_id: Slack channel ID.
        data: GitHub event data.

    Returns:
        True on success, False on errors.
    """

    # See: https://api.slack.com/methods/conversations.archive
    resp = slack.conversations_archive(channel_id)
    if not resp.ok:
        pr_url = data.pull_request.htmlurl
        msg = "State of %s is `%s`, but <#%s> can't be archived: `%s`"
        msg %= (pr_url, data.action, channel_id, resp.error)
        slack.chat_post_message(channel_id, "Failed to archive this channel")

        debug(msg)

        # TODO: Also post a reply in the log channel.

    return resp.ok

def create_channel(data, name, suffix = 1):
    """Create a public Slack channel.

    Args:
        data: GitHub event data.
        name: Desired (and valid) name of the channel.
        suffix: Optional suffix to append to the channel name.

    Returns:
        Channel ID, or "" on errors.
    """

    # Optional suffix to make the channel name unique.
    # We could add a recursion stop condition, but it's not necessary.
    n = name
    if suffix > 1:
        n += "_%d" % suffix

    # Create the channel.
    # See: https://api.slack.com/methods/conversations.create
    resp = slack.conversations_create(SLACK_CHANNEL_PREFIX + n)
    if not resp.ok:
        if resp.error == "name_taken":
            # If a channel with the same name already exists,
            # try again recursively with a numeric suffix.
            return create_channel(data, name, suffix + 1)
        else:
            debug('Create Slack channel "%s": `%s`' % (n, resp.error))
            return ""

    # As long as the channel was created, these nice-to-haves aren't critical.
    channel_id = resp.channel.id
    _set_channel_description(channel_id, data)
    _set_channel_topic(channel_id, data)

    # TODO: Post a message in the log channel.

    return channel_id

def get_permalink(channel_id, message_ts):
    """Return a markdown-formatted permalink to a specific Slack message.

    Args:
        channel_id: ID of the Slack channel containing the message.
        message_ts: Timestamp of the specific message to link to.

    Returns:
        GitHub Markdown-formatted link, or just the word "Slack"
        if we couldn't generate it.
    """

    # See: https://api.slack.com/methods/chat.getPermalink
    resp = slack.chat_get_permalink(channel_id, message_ts)
    if resp.ok:
        return "[Slack](%s)" % resp.permalink
    else:
        debug("Failed to get permalink for Slack message: `%s`" % resp.error)
        return "Slack"

def impersonate_user_in_message(channel_id, github_user, msg, github_owner_org):
    """Post a message to a Slack channel, as a user.

    See also the "mention_user_in_message" function below.

    Args:
        channel_id: ID of the channel to send the message to.
        github_user: GitHub user object of the mentioned user.
        msg: Message to send (not containing a "%s" placeholder).
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Message's thread timestamp, or "" on errors.
    """
    if not channel_id:
        return ""

    user = github_username_to_slack_user(github_user.login, github_owner_org)
    if not user:
        return ""

    # TODO: Also post the message in the log channel.
    p = user.profile

    resp = slack.chat_post_message(channel_id, msg, username = p.real_name, icon_url = p.image_48)
    return resp.ts if resp.ok else ""

def impersonate_user_in_reply(channel_id, review_url, github_user, msg, github_owner_org):
    """Post a reply to a Slack message (review comment), as a user.

    See also the "mention_user_in_reply" function below.

    Args:
        channel_id: ID of the channel to send the message to.
        review_url: URL of the GitHub PR review to comment on.
        github_user: GitHub user object of the mentioned user.
        msg: Message to send (not containing a "%s" placeholder).
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Message's thread timestamp, or "" on errors.
    """
    if not channel_id:
        return ""

    user = github_username_to_slack_user(github_user.login, github_owner_org)
    if not user:
        return ""

    # TODO: Also post the reply in the log channel.
    p = user.profile

    thread_ts = _lookup_review_message(review_url)
    if not thread_ts:
        return ""

    resp = slack.chat_post_message(channel_id, msg, thread_ts = thread_ts, username = p.real_name, icon_url = p.image_48)
    return resp.ts if resp.ok else ""

def lookup_pr_channel(pr_url, state):
    """Return the ID of a Slack channel representing a GitHub PR.

    This function waits for the channel to exist, if it doesn't already,
    up to a timeout of a few seconds. This is useful when we want to sync
    multiple events during channel creation, i.e. PR re/opening.

    Args:
        pr_url: URL of the GitHub PR.
        state: GitHub event's action.

    Returns:
        Channel ID, or "" if not found.
    """
    channel_id = lookup_github_link_details(pr_url)
    if not channel_id:
        debug("State of %s is `%s`, but Slack channel ID not found" % (pr_url, state))
    return channel_id

def _lookup_review_message(review_url):
    """Return the ID of a Slack message representing a GitHub PR review.

    This function waits for the message to exist, if it doesn't already,
    up to a timeout of a few seconds.

    Args:
        review_url: URL of the GitHub PR review to search for.

    Returns:
        Message's thread timestamp, or "" if not found.
    """
    thread_ts = lookup_github_link_details(review_url)
    if not thread_ts:
        debug("Message mapping for %s not found" % review_url)
    return thread_ts

def mention_user_in_message(channel_id, github_user, msg, github_owner_org):
    """Post a message to a Slack channel, mentioning a user.

    See also the "impersonate_user_in_message" function above.

    Args:
        channel_id: ID of the channel to send the message to.
        github_user: GitHub user object of the mentioned user.
        msg: Message to send, containing a single "%s" placeholder.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Message's thread timestamp, or "" on errors.
    """
    if not channel_id:
        return ""

    msg %= resolve_github_user(github_user, github_owner_org)

    # TODO: Also post the message in the log channel.

    resp = slack.chat_post_message(channel_id, msg)
    return resp.ts if resp.ok else ""

def mention_user_in_reply(channel_id, review_url, github_user, msg, github_owner_org):
    """Post a reply to a Slack message (review comment), mentioning a user.

    See also the "impersonate_user_in_reply" function above.

    Args:
        channel_id: ID of the channel to send the message to.
        review_url: URL of the GitHub PR review to comment on.
        github_user: GitHub user object of the mentioned user.
        msg: Message to send, containing a single "%s" placeholder.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Message's thread timestamp, or "" on errors.
    """
    if not channel_id:
        return ""

    msg %= resolve_github_user(github_user, github_owner_org)

    # TODO: Also post the reply in the log channel.

    thread_ts = _lookup_review_message(review_url)
    if not thread_ts:
        return ""

    resp = slack.chat_post_message(channel_id, msg, thread_ts = thread_ts)
    return resp.ts if resp.ok else ""

def normalize_channel_name(name):
    """Convert arbitrary text into a valid Slack channel name.

    Args:
        name: Desired name for a Slack channel.

    Returns:
        Valid Slack channel name.
    """
    name = name.lower().strip()

    # https://github.com/qri-io/starlib/tree/master/re
    name = re.sub(r"'\"", "", name)
    name = re.sub(r"[^a-z0-9_-]", "-", name)
    name = re.sub(r"[_-]{2,}", "-", name)

    # Slack channel names are limited to 80 characters, but that's
    # too long for comfort, so we use 50 instead. Plus, we need to
    # leave room for a PR number prefix and a uniqueness suffix.
    name = name[:50]

    # Cosmetic tweak: remove leading and trailing hyphens.
    if name[0] == "-":
        name = name[1:]
    if name[-1] == "-":
        name = name[:-1]

    return name

def rename_channel(channel_id, name, suffix = 1):
    """Rename a Slack channel.

    Args:
        channel_id: Slack channel ID.
        name: Desired (and valid) name of the channel.
        suffix: Optional suffix to append to the channel name.
    """

    # Optional suffix to make the channel name unique.
    # We could add a recursion stop condition, but it's not necessary.
    n = name
    if suffix > 1:
        n += "_%d" % suffix

    # Rename the channel.
    # See: https://api.slack.com/methods/conversations.rename
    resp = slack.conversations_rename(channel_id, SLACK_CHANNEL_PREFIX + n)
    if not resp.ok:
        if resp.error == "name_taken":
            # If a channel with the same name already exists,
            # try again recursively with a numeric suffix.
            rename_channel(channel_id, name, suffix + 1)
            return
        else:
            debug('Rename Slack channel to "%s": `%s`' % (n, resp.error))
            return

def _set_channel_description(channel_id, data):
    """Set the description of a Slack channel to a GitHub PR title.

    Args:
        channel_id: Slack channel ID.
        data: GitHub event data.
    """
    pr = data.pull_request
    s = "`%s`" % pr.title
    if len(s) > _CHANNEL_MAX_METADATA_LENGTH:
        s = s[:_CHANNEL_MAX_METADATA_LENGTH - 4] + "`..."

    # See: https://api.slack.com/methods/conversations.setPurpose
    resp = slack.conversations_set_purpose(channel_id, s)
    if not resp.ok:
        msg = "State of %s is `%s`, but <#%s> can't be updated: `%s`"
        debug(msg % (pr.htmlurl, data.action, channel_id, resp.error))

def _set_channel_topic(channel_id, data):
    """Set the topic of a Slack channel to a GitHub PR URL.

    Args:
        channel_id: Slack channel ID.
        data: GitHub event data.
    """
    pr = data.pull_request
    s = pr.htmlurl
    if len(s) > _CHANNEL_MAX_METADATA_LENGTH:
        s = s[:_CHANNEL_MAX_METADATA_LENGTH - 4] + " ..."

    # See: https://api.slack.com/methods/conversations.setTopic
    resp = slack.conversations_set_topic(channel_id, s)
    if not resp.ok:
        msg = "State of %s is `%s`, but <#%s> can't be updated: `%s`"
        debug(msg % (pr.htmlurl, data.action, channel_id, resp.error))

def unarchive_channel(channel_id, data):
    """Unarchive a Slack channel.

    Attention - https://api.slack.com/methods/conversations.unarchive:
    Bug alert: bot tokens (xoxb-...) cannot currently be used to unarchive
    conversations. For now, please use a user token (xoxp-...) to unarchive
    the conversation rather than a bot token.

    Args:
        channel_id: Slack channel ID.
        data: GitHub event data.

    Returns:
        True on success, False on errors.
    """

    # See: https://api.slack.com/methods/conversations.unarchive
    resp = slack.conversations_unarchive(channel_id)
    if not resp.ok:
        pr_url = data.pull_request.htmlurl
        msg = "State of %s is `%s`, but <#%s> can't be unarchived: `%s`"
        debug(msg % (pr_url, data.action, channel_id, resp.error))

    # TODO: Also post a reply in the log channel.

    return resp.ok
