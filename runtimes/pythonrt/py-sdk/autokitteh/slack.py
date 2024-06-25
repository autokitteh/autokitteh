"""Initialize a Slack client, based on an AutoKitteh connection."""

import os

from slack_sdk.web.client import WebClient

from .connections import check_connection_name
from .errors import ConnectionInitError


def slack_client(connection: str, **kwargs) -> WebClient:
    """Initialize a Slack client, based on an AutoKitteh connection.

    API reference:
    https://slack.dev/python-slack-sdk/api-docs/slack_sdk/web/client.html

    This function doesn't initialize a Socket Mode client because the
    AutoKitteh connection already has one to receive incoming events.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Slack SDK client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        SlackApiError: Connection attempt failed, or connection is unauthorized.
    """
    check_connection_name(connection)

    bot_token = os.getenv(connection + "__oauth_AccessToken")  # OAuth v2
    if not bot_token:
        bot_token = os.getenv(connection + "__BotToken")  # Socket Mode
    if not bot_token:
        raise ConnectionInitError(connection)

    client = WebClient(bot_token, **kwargs)
    client.auth_test().validate()
    return client
