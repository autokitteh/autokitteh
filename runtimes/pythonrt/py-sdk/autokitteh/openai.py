"""Initialize an OpenAI client, based on an AutoKitteh connection."""

import os

from openai import OpenAI

from .errors import ConnectionInitError


def openai_client(connection: str) -> OpenAI:
    """Initialize an OpenAI client, based on an AutoKitteh connection.

    API reference:
    https://platform.openai.com/docs/api-reference/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        OpenAI API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OpenAIError: Connection attempt failed, or connection is unauthorized.
    """
    api_key = os.getenv(connection + "__api_key")
    if not api_key:
        raise ConnectionInitError(connection)

    return OpenAI(api_key=api_key)
