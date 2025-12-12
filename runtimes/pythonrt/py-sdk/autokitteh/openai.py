"""Initialize an OpenAI client, based on an AutoKitteh connection."""

import os

from openai import OpenAI
from pydantic_ai.providers.openai import OpenAIProvider

from .connections import check_connection_name
from .errors import ConnectionInitError


def openai_client(connection: str) -> OpenAI:
    """Initialize an OpenAI client, based on an AutoKitteh connection.

    API reference:
    https://platform.openai.com/docs/api-reference/
    https://github.com/openai/openai-python/blob/main/api.md

    Args:
        connection: AutoKitteh connection name.

    Returns:
        OpenAI API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OpenAIError: Connection attempt failed, or connection is unauthorized.
    """
    check_connection_name(connection)

    api_key = os.getenv(connection + "__apiKey")

    if not api_key:
        raise ConnectionInitError(connection)

    return OpenAI(api_key=api_key)


def openai_pydantic_ai_provider(connection: str, **kwargs) -> OpenAIProvider:
    """Initialize an OpenAI Pydantic AI provider, based on an AutoKitteh connection.

    API reference:
        https://ai.pydantic.dev/models/openai

    Args:
        connection: AutoKitteh connection name.

    Returns:
        OpenAI Pydantic AI provider.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OpenAIError: Connection attempt failed, or connection is unauthorized.
    """
    check_connection_name(connection)

    api_key = os.getenv(connection + "__apiKey")

    if not api_key:
        raise ConnectionInitError(connection)

    return OpenAIProvider(api_key=api_key, **kwargs)
