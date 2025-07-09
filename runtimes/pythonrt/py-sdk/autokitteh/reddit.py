"""Initializes a Reddit client, based on an AutoKitteh connection."""

import os

import praw

from .connections import check_connection_name
from .errors import ConnectionInitError


def reddit_client(connection: str) -> praw.Reddit:
    """Initialize a Reddit client, based on an AutoKitteh connection.

    API reference:
    https://praw.readthedocs.io/en/stable/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Reddit API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    client_id = os.getenv(connection + "__client_id")
    client_secret = os.getenv(connection + "__client_secret")
    user_agent = os.getenv(connection + "__user_agent")
    username = os.getenv(connection + "__username")
    password = os.getenv(connection + "__password")

    if not client_id or not client_secret or not user_agent:
        raise ConnectionInitError(connection)

    if username and password:
        return praw.Reddit(
            client_id=client_id,
            client_secret=client_secret,
            user_agent=user_agent,
            username=username,
            password=password,
        )

    # read-only client.
    return praw.Reddit(
        client_id=client_id,
        client_secret=client_secret,
        user_agent=user_agent,
    )
