"""Handler for GitHub "pull_request_review_thread" events."""

load("debug.star", "debug")

def on_github_pull_request_review_thread(data):
    """https://docs.github.com/webhooks/webhook-events-and-payloads#pull_request_review_thread

    Args:
        data: GitHub event data.
    """
    action_handlers = {
        "resolved": _on_pr_review_thread_resolved,
        "unresolved": _on_pr_review_thread_unresolved,
    }
    if data.action in action_handlers:
        action_handlers[data.action](data)
    else:
        debug("Unrecognized GitHub PR review thread action: `%s`" % data.action)

def _on_pr_review_thread_resolved(data):
    """A comment thread on a pull request was marked as resolved.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.thread)
    print(data.sender)
    print(data.pull_request)

def _on_pr_review_thread_unresolved(data):
    """A previously resolved comment thread on a pull request was marked as unresolved.

    TODO: Implement this.

    Args:
        data: GitHub event data.
    """
    print(data.thread)
    print(data.sender)
    print(data.pull_request)
