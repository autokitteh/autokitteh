"""AutoKitteh Python SDK."""

from datetime import UTC, datetime
import os
import re

from atlassian import Jira
import slack_sdk

from uuid import uuid4


ACTIVITY_ATTR = "__activity__"

class AttrDict(dict):
    """Allow attribute access to dictionary keys.

    >>> config = AttrDict({'server': {'port': 8080}, 'debug': True})
    >>> config.server.port
    8080
    >>> config.debug
    True
    """
    def __getattr__(self, name):
        try:
            value = self[name]
            if isinstance(value, dict):
                value = AttrDict(value)
            return value
        except KeyError:
            raise AttributeError(name)

    def __setattr__(self, attr, value):
        # The default __getattr__ doesn't fail but also don't change values
        cls = self.__class__.__name__
        raise NotImplementedError(f'{cls} does not support setting attributes')


def activity(fn: callable) -> callable:
    """Decorator to mark a function as an activity."""
    setattr(fn, ACTIVITY_ATTR, True)
    return fn


<<<<<<< HEAD
def subscribe(connection_name: str, filter: str) -> str:
    """Subscribe to events on connection. Option filter is a CEL expression."""
=======
def jira_client(connection: str, **kwargs) -> Jira:
    """Initialize a Jira client, based on an AutoKitteh connection.

    API reference:
    https://atlassian-python-api.readthedocs.io/jira.html

    Code examples:
    https://github.com/atlassian-api/atlassian-python-api/tree/master/examples/jira

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Atlassian-Python-API Jira client.
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError("Invalid AutoKitteh connection name: " + connection)

    if os.getenv(connection + "__oauth_AccessToken"):
        return _jira_client_cloud_oauth2(connection, **kwargs)

    raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')


def _jira_client_cloud_oauth2(connection: str, **kwargs) -> Jira:
    """Initialize a Jira client for Atlassian Cloud using OAuth 2.0."""
    expiry = os.getenv(connection + "__oauth_Expiry")
    if not expiry:
        raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')

    iso8601 = re.sub(r"[ A-Z]+$", "", expiry)  # Convert from Go's time string.
    if datetime.fromisoformat(iso8601) < datetime.now(UTC):
        raise RuntimeError("OAuth 2.0 access token expired on: " + expiry)

    cloud_id = os.getenv(connection + "__access_id")
    if not cloud_id:
        raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')

    client_id = os.getenv("JIRA_CLIENT_ID")
    if not client_id:
        raise RuntimeError('Environment variable "JIRA_CLIENT_ID" not set')

    return Jira(
        url="https://api.atlassian.com/ex/jira/" + cloud_id,
        oauth2={
            "client_id": client_id,
            "token": {
                "access_token": os.getenv(connection + "__oauth_AccessToken"),
                "token_type": os.getenv(connection + "__oauth_TokenType"),
            },
        },
        **kwargs,
    )


def slack_client(connection: str, **kwargs) -> slack_sdk.web.client.WebClient:
    """Initialize a Slack client, based on an AutoKitteh connection.
>>>>>>> cce9af3b (Jira with OAuth 2.0)

    # Dummy implementation for local development
    return uuid4().hex


def unsubscribe(id: str) -> None:
    """Unsubscribe from events."""

<<<<<<< HEAD
    # Dummy implementation for local development
    pass
=======
    Returns:
        Slack SDK client.
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError("Invalid AutoKitteh connection name: " + connection)

    bot_token = os.getenv(connection + "__oauth_AccessToken")  # OAuth v2
    if not bot_token:
        bot_token = os.getenv(connection + "__BotToken")  # Socket Mode
    if not bot_token:
        raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')
>>>>>>> cce9af3b (Jira with OAuth 2.0)


def next_event(id: str) -> AttrDict:
    """Get the next event from the subscription."""

    # Dummy implementation for local development
    return AttrDict()
