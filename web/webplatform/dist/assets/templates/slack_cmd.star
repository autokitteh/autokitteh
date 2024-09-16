"""Handler for Slack slash-command events."""

load("@slack", "slack")
load(
    "redis_helpers.star",
    "del_slack_opt_out",
    "get_slack_opt_out",
    "set_slack_opt_out",
)

WAKE_WORD = "purrr"
HELP_CMD = "help"

def on_slack_slash_command(data):
    """https://api.slack.com/interactivity/slash-commands

    See also: https://api.slack.com/interactivity/handling#message_responses

    Args:
        data: Slack event data.
    """

    # Split the command string into normalized arguments.
    # See: https://github.com/qri-io/starlib/tree/master/re
    args = re.split(r"\s+", data.text.lower().strip())

    # No command? Nothing to do.
    if len(args) == 0:
        return

    # Just "help"? Only hint at a more specific help command (i.e. "help"
    # with our wake-word), so we won't interfere with other autokitteh
    # projects that may be reusing the same Slack connection token to
    # implement their own slash commands with their own help messages.
    if args == (HELP_CMD,):
        # Why an ephemeral message? No need to spam the channel,
        # show the help message only to the user who asked for it.
        text = "Type: `%s %s %s`" % (data.command, HELP_CMD, WAKE_WORD)
        slack.chat_post_ephemeral(data.channel_id, data.user_id, text)
        return

    # Do nothing if the command doesn't start with our wake-word (it's probably
    # being handled by a different autokitteh project that is reusing the same
    # Slack connection token), unless it's "help" followed by our wake-word.
    if WAKE_WORD in args and (HELP_CMD in args or len(args) == 1):
        _help(data, args)
        return
    if args[0] != WAKE_WORD:
        return

    # Route further processing to the appropriate sub-command handler.
    for cmd, _, func in COMMANDS:
        if cmd == args[1]:
            func(data, args[1:])
            return

    error = "Error: unrecognized `%s` sub-command: `%s`" % (WAKE_WORD, args[1])
    slack.chat_post_ephemeral(data.channel_id, data.user_id, error)

def _error(data, cmd, msg):
    error = "Error in `%s %s %s`: %s" % (data.command, WAKE_WORD, cmd, msg)
    slack.chat_post_ephemeral(data.channel_id, data.user_id, error)

def _help(data, args):
    """Help command.

    Args:
        data: Slack event data.
        args: Tuple of normalized string arguments.
    """
    if len(args) > 2:
        _error(data, HELP_CMD, "this command doesn't accept extra arguments")
        return

    help = ":wave: *GitHub Pull Request Review Reminder (PuRRR)* :wave:\n\nAvailable slash commands:"
    for cmd, description, _ in COMMANDS:
        help += "\n  •  `%s %s %s` - %s" % (data.command, WAKE_WORD, cmd, description)
    slack.chat_post_ephemeral(data.channel_id, data.user_id, help)

def _opt_in(data, args):
    """Opt-in command.

    Args:
        data: Slack event data.
        args: Tuple of normalized string arguments.
    """
    if len(args) > 1:
        _error(data, args[0], "this command doesn't accept extra arguments")
        return

    if not get_slack_opt_out(data.user_id):
        msg = ":bell: You're already opted into PuRRR"
        slack.chat_post_ephemeral(data.channel_id, data.user_id, msg)
        return

    del_slack_opt_out(data.user_id)
    msg = ":bell: You are now opted into PuRRR"
    slack.chat_post_ephemeral(data.channel_id, data.user_id, msg)

def _opt_out(data, args):
    """Opt-out command.

    Args:
        data: Slack event data.
        args: Tuple of normalized string arguments.
    """
    if len(args) > 1:
        _error(data, args[0], "this command doesn't accept extra arguments")
        return

    opt_out_time = get_slack_opt_out(data.user_id)
    if opt_out_time:
        msg = ":no_bell: You're already opted out of PuRRR since: " + opt_out_time
        slack.chat_post_ephemeral(data.channel_id, data.user_id, msg)
        return

    if set_slack_opt_out(data.user_id):
        msg = ":no_bell: You are now opted out of PuRRR"
        slack.chat_post_ephemeral(data.channel_id, data.user_id, msg)

def _list(data, args):
    """List command.

    Args:
        data: Slack event data.
        args: Tuple of normalized string arguments.
    """
    if len(args) > 1:
        _error(data, args[0], "this command doesn't accept extra arguments")
        return

    slack.chat_post_ephemeral(data.channel_id, data.user_id, "TODO: implement me!")

def _status(data, args):
    """Status command.

    Args:
        data: Slack event data.
        args: Tuple of normalized string arguments.
    """

    # TODO: If the Slack channel belongs to a PR, the args are optional.
    if len(args) != 2:
        msg = "this command requires exactly 1 argument - an ID of a "
        msg += "GitHub PR (`<org>/<repo>/<number>`), or its full URL"
        _error(data, args[0], msg)
        return

    slack.chat_post_ephemeral(data.channel_id, data.user_id, "TODO: implement me!")

def _approve(data, args):
    """Approve command.

    Args:
        data: Slack event data.
        args: Tuple of normalized string arguments.
    """

    # TODO: If the Slack channel belongs to a PR, the args are optional.
    if len(args) != 2:
        msg = "this command requires exactly 1 argument - an ID of a "
        msg += "GitHub PR (`<org>/<repo>/<number>`), or its full URL"
        _error(data, args[0], msg)
        return

    slack.chat_post_ephemeral(data.channel_id, data.user_id, "TODO: implement me!")

COMMANDS = [
    ("opt-in", "Opt into receiving notifications", _opt_in),
    ("opt-out", "Opt out of receiving notifications", _opt_out),
    ("list", "List all PRs you should pay attention to", _list),
    ("status [PR]", "Check the status of a specific PR", _status),
    ("approve [PR]", "Approve a specific PR", _approve),
]
