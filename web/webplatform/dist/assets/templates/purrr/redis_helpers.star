"""Redis API helper functions."""

load("@redis", "redis")
load("debug.star", "debug")
load("env", "REDIS_TTL")  # Set in "autokitteh.yaml".

_GET_TIMEOUT = 5  # Seconds.

# Optimization: cache user lookup results for a day, to
# reduce the amount of API calls, especially to Slack.
_USER_CACHE_TTL = "24h"

def _del(key):
    """https://redis.io/commands/del/"""
    redis.delete(key)

def _get(key, wait):
    """https://redis.io/commands/get/ + optional retries"""
    attempts = _GET_TIMEOUT if wait else 1
    for _ in range(attempts):
        value = redis.get(key)
        if value:
            return value
        else:
            sleep(1)  # Wait for the key to exist, up to a point.

    # Timeout.
    return ""

def _set(key, value, ttl = None):
    """https://redis.io/commands/set/"""
    if ttl:
        resp = redis.set(key, value, ttl)
    else:
        resp = redis.set(key, value)
    if resp != "OK":
        debug("Redis `set %s %s %s` failed: `%s`" % (key, value, ttl, resp))
    return resp == "OK"

def cache_github_reference(slack_user_id, github_ref):
    """Optimization to reduce the amount of API calls in "user_helpers.star"."""
    _set("slack_user:" + slack_user_id, github_ref, _USER_CACHE_TTL)

def cached_github_reference(slack_user_id):
    """Optimization to reduce the amount of API calls in "user_helpers.star".

    Args:
        slack_user_id: Slack user ID to look-up.

    Returns:
        GitHub user reference ("@username"), the Slack user's full name, or "" if not found.
    """
    github_ref = _get("slack_user:" + slack_user_id, wait = False)
    if github_ref:
        # Optimization: extend the TTL after a successful cache hit.
        # See: https://redis.io/commands/expire/
        redis.expire("slack_user:" + slack_user_id, _USER_CACHE_TTL)

    return github_ref

def cache_slack_user_id(github_username, slack_user_id):
    """Optimization to reduce the amount of API calls in "user_helpers.star"."""
    _set("github_user:" + github_username, slack_user_id, _USER_CACHE_TTL)

def cached_slack_user_id(github_username):
    """Optimization to reduce the amount of API calls in "user_helpers.star".

    Args:
        github_username: GitHub username to look-up.

    Returns:
        Slack user ID, or "" if not found.
    """
    slack_user_id = _get("github_user:" + github_username, wait = False)
    if slack_user_id:
        # Optimization: extend the TTL after a successful cache hit,
        # but only for bots and real users. If the cached result is
        # "not found" then reevaluate it on a daily basis.
        # See: https://redis.io/commands/expire/
        if slack_user_id != "not found":
            redis.expire("github_user:" + github_username, _USER_CACHE_TTL)

    return slack_user_id

def map_github_link_to_slack_channel_id(github_link, slack_channel_id):
    """Called in "github_pr.star", used by future GitHub events."""
    _set(github_link, slack_channel_id, REDIS_TTL)

def lookup_github_link_details(github_link):
    return _get(github_link, wait = True)

def map_slack_channel_id_to_pr_details(slack_channel_id, org, repo, pr_number):
    """Called in "github_pr.star", used by future Slack events."""
    _set(slack_channel_id, "%s:%s:%s" % (org, repo, pr_number), REDIS_TTL)

def translate_slack_channel_id_to_pr_details(slack_channel_id):
    """Synchronize Slack events to GitHub PRs."""
    pr_details = _get(slack_channel_id, wait = True) or "::0"
    owner, repo, pr = pr_details.split(":")
    return owner, repo, int(pr)

def map_github_link_to_slack_message_ts(github_link, slack_message_ts):
    _set(github_link, slack_message_ts, REDIS_TTL)

def map_slack_message_ts_to_github_link(slack_message_ts, github_link):
    _set(slack_message_ts, github_link, REDIS_TTL)

def translate_slack_review_comment_to_github_id(channel_id, message_ts):
    """Called by create_review_comment_reply() in "github_helpers.star"."""
    id = _get("review_comment:%s:%s" % (channel_id, message_ts), wait = True) or "0"
    return int(id)

def translate_slack_message_to_github_link(message_type, channel_id, message_ts):
    """Called by create_review_comment_reply() in "github_helpers.star"."""
    return _get("%s:%s:%s" % (message_type, channel_id, message_ts), wait = False)

def del_slack_opt_out(slack_user_id):
    """Called by _opt_in() in "slack_cmd.star"."""
    _del("slack_opt_out:" + slack_user_id)

def get_slack_opt_out(slack_user_id):
    return _get("slack_opt_out:" + slack_user_id, wait = False)

def set_slack_opt_out(slack_user_id):
    """Called by _opt_out() in "slack_cmd.star"."""
    return _set("slack_opt_out:" + slack_user_id, time.now())  # No expiration.
