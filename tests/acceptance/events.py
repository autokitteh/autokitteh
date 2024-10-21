"""See https://docs.autokitteh.com/develop/events/types"""

import api_calls
import print


def on_http_request(event):
    """To trigger this function, run this command:
    curl -i "http[s]://autokitteh-address/webhooks/trigger-slug"
    """
    print.pretty_json("HTTP trigger event", event)


def on_github_issue_comment(event):
    """To trigger this function, add/edit/delete a comment on a GitHub
    issue in a repository that the AutoKitteh app has been installed in.
    """
    print.pretty_json("GitHub issue comment event", event)
    api_calls.github_get_repo(event)


def on_slack_message(event):
    """To trigger this function, post a message in your Slack DM with the
    AutoKitteh app, or a private/public channel that it has been added to.
    """
    print.pretty_json("Slack message event", event)
    api_calls.slack_auth_test(event)
