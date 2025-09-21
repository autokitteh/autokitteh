"""Initialize Notion client, based on an AutoKitteh connection."""

import os

from notion_client import Client

from .connections import check_connection_name
from .errors import ConnectionInitError


def notion_client(connection: str) -> Client:
    """Initialize a Notion client, based on an AutoKitteh connection.

    API reference:
        https://developers.notion.com/docs
        https://github.com/ramnes/notion-sdk-py

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Notion client object.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    notion_secret = os.getenv(f"{connection}__internal_integration_secret")

    if notion_secret:
        return Client(auth=notion_secret)

    access_token = os.getenv(connection + "__oauth_access_token")
    if not access_token:
        raise ConnectionInitError(connection)

    return Client(auth=access_token)
