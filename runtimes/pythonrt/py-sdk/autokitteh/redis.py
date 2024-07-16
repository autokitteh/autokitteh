"""Initialize a Redis client, based on an AutoKitteh connection."""

import os

from redis import Redis

from .connections import check_connection_name
from .errors import ConnectionInitError


def redis_client(connection: str, **kwargs) -> Redis:
    """Initialize a Redis client, based on an AutoKitteh connection.

    API reference and examples: https://redis-py.readthedocs.io/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Redis client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection's URL not configured.
    """
    check_connection_name(connection)

    url = os.getenv(connection + "__url")
    if not url:
        raise ConnectionInitError(connection)

    return Redis.from_url(url, **kwargs)
