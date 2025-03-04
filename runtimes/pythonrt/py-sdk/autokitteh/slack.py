"""Slack client initialization, and other helper functions."""

import os
import re

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

    bot_token = os.getenv(connection + "__oauth_access_token")  # OAuth v2
    if not bot_token:
        bot_token = os.getenv(connection + "__private_bot_token")  # Socket Mode
    if not bot_token:  # TODO(INT-267): Remove this old env var check.
        bot_token = os.getenv(connection + "__oauth_AccessToken")
    if not bot_token:  # TODO(INT-267): Remove this old env var check.
        bot_token = os.getenv(connection + "__BotToken")
    if not bot_token:
        raise ConnectionInitError(connection)

    client = WebClient(bot_token, **kwargs)
    client.auth_test().validate()
    return client


def normalize_channel_name(name: str) -> str:
    """Convert arbitrary text into a valid Slack channel name.

    See: https://api.slack.com/methods/conversations.create#naming

    Args:
        name: Desired name for a Slack channel.

    Returns:
        Valid Slack channel name.
    """
    if name == "":
        return name

    name = name.lower().strip()
    name = re.sub(r"['\"]", "", name)  # Remove quotes.
    name = re.sub(r"[^a-z0-9_-]", "-", name)  # Replace invalid characters.
    name = re.sub(r"[_-]{2,}", "-", name)  # Remove repeating separators.

    # Slack channel names are limited to 80 characters,
    # but that's too long for comfort, so we use 50 instead.
    name = name[:50]

    # Cosmetic tweak: remove leading and trailing hyphens.
    if name[0] == "-":
        name = name[1:]
    if name[-1] == "-":
        name = name[:-1]

    return name
