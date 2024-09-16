"""User-related helper functions across GitHub and Slack."""

load("@github", "github")
load("@slack", "slack")
load("debug.star", "debug")
load(
    "redis_helpers.star",
    "cache_github_reference",
    "cache_slack_user_id",
    "cached_github_reference",
    "cached_slack_user_id",
)

def _email_to_github_user_id(email, github_owner_org):
    """Convert an email address into a GitHub user ID.

    Args:
        email: User's email address.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        GitHub user ID, or "" if not found.
    """

    # See: https://docs.github.com/en/rest/search/search#search-users
    # And: https://docs.github.com/en/search-github/searching-on-github/searching-users
    resp = github.search_users(email + " in:email", owner = github_owner_org)
    if resp.total == 1:
        return resp.users[0].login
    else:
        debug("GitHub search results: %d users with the email address `%s`" % (resp.total, email))
        return ""

def _email_to_slack_user_id(email):
    """Convert an email address into a Slack user ID.

    Args:
        email: Email address.

    Returns:
        Slack user ID, or "" if not found.
    """

    # See: https://api.slack.com/methods/users.lookupByEmail
    resp = slack.users_lookup_by_email(email)
    if resp.ok:
        return resp.user.id
    else:
        debug("Look-up Slack user by email %s: `%s`" % (email, resp.error))
        return ""

def github_pr_participants(pr):
    """Return all the participants in the given GitHub pull request.

    Args:
        pr: GitHub pull request object.

    Returns:
        List of usernames (author/reviewers/assignees),
        guaranteed to be sorted and without repetitions.
    """
    usernames = []

    # Author.
    if pr.user.type == "User":
        usernames.append(pr.user.login)

    # Specific reviewers (not reviewing teams) + assignees.
    for user in pr.requested_reviewers + pr.assignees:
        if user.type == "User" and user.login not in usernames:
            usernames.append(user.login)

    return sorted(usernames)

def github_username_to_slack_user(username, github_owner_org):
    """Convert a GitHub username into a Slack user object.

    Args:
        username: GitHub username.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Slack user object, or None if not found.
    """
    slack_user_id = github_username_to_slack_user_id(username, github_owner_org)
    if not slack_user_id:
        return None

    # See: https://api.slack.com/methods/users.info
    resp = slack.users_info(slack_user_id)
    if not resp.ok:
        debug("Get Slack user info for <@%s>: `%s`" % (slack_user_id, resp.error))
        return None

    return resp.user

def github_username_to_slack_user_id(github_username, github_owner_org):
    """Convert a GitHub username into a Slack user ID.

    This function tries to match the email address first, and then
    falls back to matching the user's full name (case-insensitive).

    This function also caches successful results for a day,
    to reduce the amount of API calls, especially to Slack.

    Args:
        github_username: GitHub username.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Slack user ID, or "" if not found.
    """

    # Optimization: if we already have it cached, no need to look it up.
    slack_user_id = cached_slack_user_id(github_username)
    if slack_user_id:
        if slack_user_id in ("bot", "not found"):
            slack_user_id = ""
        return slack_user_id

    # See: https://docs.github.com/en/rest/users#get-a-user
    resp = github.get_user(github_username, owner = github_owner_org)
    github_user_link = "<%s|%s>" % (resp.htmlurl, github_username)

    # Special case: GitHub bots can't have Slack identities.
    if resp.type == "Bot":
        cache_slack_user_id(github_username, "bot")
        return ""

    # Try to match by the email address first.
    if not resp.email:
        debug("GitHub user %s: email address not found" % github_user_link)
    else:
        slack_user_id = _email_to_slack_user_id(resp.email)
        if slack_user_id:
            cache_slack_user_id(github_username, slack_user_id)
            return slack_user_id

    # Otherwise, try to match by the user's full name.
    if not resp.name:
        debug("GitHub user %s: full name not found" % github_user_link)
        return ""

    gh_full_name = resp.name.lower()
    for user in _slack_users():
        slack_names = (
            user.profile.real_name.lower(),
            user.profile.real_name_normalized.lower(),
        )
        if gh_full_name in slack_names:
            cache_slack_user_id(github_username, user.id)
            return user.id

    # Optimization: cache unsuccessful results too (i.e. external users).
    debug("GitHub user %s: email & name not found in Slack" % github_user_link)
    cache_slack_user_id(github_username, "not found")
    return ""

def resolve_github_user(github_user, github_owner_org):
    """Convert a GitHub username to a linkified user reference in Slack.

    Args:
        github_user: GitHub user object.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Slack user reference, or GitHub profile link.
        Used for mentioning users in Slack messages.
    """
    id = github_username_to_slack_user_id(github_user.login, github_owner_org)
    if id:
        # Mention the user by their Slack ID, if possible.
        return "<@%s>" % id
    else:
        # Otherwise, fall-back to their GitHub profile link.
        return "<%s|%s>" % (github_user.htmlurl, github_user.login)

def resolve_slack_user(slack_user_id, github_owner):
    """Convert a Slack user ID to a GitHub user reference.

    This function also caches successful results for a day,
    to reduce the amount of API calls, especially to Slack.

    Args:
        slack_user_id: Slack user ID.
        github_owner: Required for GitHub org-specific visibility.

    Returns:
        GitHub user reference, or the Slack user's full name, or "Someone".
        Used for mentioning users in GitHub reviews and comments.
    """
    if not slack_user_id:
        debug("Slack user ID not found in Slack message event")
        return "Someone"

    # Optimization: if we already have it cached, no need to look it up.
    github_ref = cached_github_reference(slack_user_id)
    if github_ref:
        return github_ref

    # See: https://api.slack.com/methods/users.info
    resp = slack.users_info(slack_user_id)
    if not resp.ok:
        debug("Get Slack user info for <@%s>: `%s`" % (slack_user_id, resp.error))
        return "Someone"

    # Special case: Slack bots can't have GitHub identities.
    if resp.user.is_bot:
        bot_name = resp.user.real_name + " (Slack bot)"
        cache_github_reference(slack_user_id, bot_name)
        return bot_name

    # Try to match by the email address first.
    email = getattr(resp.user.profile, "email", "")  # May be None.
    if not email:
        debug("Slack user <@%s>: email address not found" % slack_user_id)
    else:
        github_id = _email_to_github_user_id(email, github_owner)
        if github_id:
            github_ref = "@" + github_id
            cache_github_reference(slack_user_id, github_ref)
            return github_ref

    # TODO: Otherwise, try to match by the user's full name?
    # (Unlike Slack, where we limit the user list to a specific workspace,
    # this would search across all GitHub users, which is risky and inefficient).

    # Otherwise, return the user's full name.
    return resp.user.real_name

def _slack_users(cursor = ""):
    """Return a list of all Slack users in the workspace.

    This function uses recursion for pagination because
    Starlark doesn't officially support the "while" statement
    (even though autokitteh does, with starlark-go).

    Args:
        cursor: Optional, for pagination (initial value must be "").

    Returns:
        List of all Slack users in the workspace.
    """

    # See: https://api.slack.com/methods/users.list
    resp = slack.users_list(cursor, limit = 100)
    if not resp.ok:
        debug("List Slack users (cursor `%s`): `%s`" % (cursor, resp.error))
        return []

    users = resp.members
    if resp.response_metadata.next_cursor:
        users += _slack_users(resp.response_metadata.next_cursor)
    return users
