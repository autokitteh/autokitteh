"""Initialize an OpenAI client, based on an AutoKitteh connection."""

import os

from pydantic_ai.providers.anthropic import AnthropicProvider
from pydantic_ai.providers.gateway import gateway_provider
from pydantic_ai.providers.openai import OpenAIProvider

from .connections import check_connection_name
from .errors import ConnectionInitError


def pydantic_gateway_provider(connection: str, *args, **kwargs) -> gateway_provider:
    """Initialize a Pydantic Gateway provider, based on an AutoKitteh connection.

    API reference:
    https://platform.openai.com/docs/api-reference/
    https://github.com/openai/openai-python/blob/main/api.md

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Pydantic Gateway provider.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    api_key = os.getenv(connection + "__apiKey")

    if not api_key:
        raise ConnectionInitError(connection)

    return gateway_provider(api_key=api_key, *args, **kwargs)


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


def anthropic_pydantic_ai_provider(connection: str, **kwargs) -> AnthropicProvider:
    """Initialize an Anthropic Pydantic AI provider, based on an AutoKitteh connection.

    API reference:
        https://ai.pydantic.dev/models/anthropic

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Anthropic Pydantic AI provider.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        anthropic.APIError: Anthropic SDK initialization errors.
    """
    check_connection_name(connection)

    api_key = os.getenv(connection + "__api_key")

    if not api_key:
        raise ConnectionInitError(connection)

    return AnthropicProvider(api_key=api_key, **kwargs)
