"""Initialize an Anthropic client, based on an AutoKitteh connection."""

import os

from anthropic import Anthropic

from .connections import check_connection_name
from .errors import ConnectionInitError


def anthropic_client(connection: str) -> Anthropic:
    """Initialize an Anthropic client, based on an AutoKitteh connection.

    API reference:
    https://docs.anthropic.com/claude/reference
    https://github.com/anthropics/anthropic-sdk-python

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Anthropic API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        anthropic.APIError: Anthropic SDK initialization errors.
    """
    check_connection_name(connection)

    api_key = os.getenv(connection + "__api_key")

    if not api_key:
        raise ConnectionInitError(connection)

    return Anthropic(api_key=api_key)
