import os

import asana

from .connections import check_connection_name
from .errors import ConnectionInitError


def asana_client(connection: str) -> asana.ApiClient:
    """Initialize an Asana client, based on an AutoKitteh connection.

    API reference:
    https://developers.asana.com/docs/python

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Asana client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    access_token = os.getenv(connection + "__pat")
    if not access_token:
        raise ConnectionInitError(connection)

    configuration = asana.Configuration()
    configuration.access_token = access_token
    return asana.ApiClient(configuration)
