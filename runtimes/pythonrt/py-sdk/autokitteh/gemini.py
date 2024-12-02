"""Initialize a Gemini generative AI client, based on an AutoKitteh connection."""

import os

import google.generativeai as genai

from .connections import check_connection_name
from .errors import ConnectionInitError


def gemini_client(connection: str, **kwargs) -> genai.GenerativeModel:
    """Initialize a genai client, based on an AutoKitteh connection.

    API reference:
    - https://ai.google.dev/gemini-api/docs
    - https://github.com/google-gemini/generative-ai-python/blob/main/docs/api/google/generativeai/GenerativeModel.md

    Code samples:
    - https://ai.google.dev/gemini-api/docs#explore-the-api
    - https://ai.google.dev/gemini-api/docs/text-generation?lang=python
    - https://github.com/google-gemini/generative-ai-python/tree/main/samples
    - https://github.com/google-gemini/cookbook

    Args:
        connection: AutoKitteh connection name.

    Returns:
        An initialized GenerativeModel instance.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.

    """
    check_connection_name(connection)

    # Set the API key, if possible.
    api_key = os.getenv(connection + "__api_key")
    if not api_key:
        raise ConnectionInitError(connection)

    genai.configure(api_key=api_key)
    return genai.GenerativeModel(**kwargs)
