"""Initialize an Airtable client, based on an AutoKitteh connection."""

import os

from pyairtable import Api

from .connections import check_connection_name, refresh_oauth
from .errors import ConnectionInitError, OAuthRefreshError


def airtable_client(connection: str) -> Api:
    """Initialize an Airtable client, based on an AutoKitteh connection.

    API reference:
    https://pyairtable.readthedocs.io/en/stable/getting-started.html
    https://github.com/gtalarico/pyairtable

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Requests session object.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    check_connection_name(connection)

    pat = os.getenv(f"{connection}__pat")
    if pat:
        return Api(pat)

    refresh_token = os.getenv(connection + "__oauth_refresh_token")
    if not refresh_token:
        raise ConnectionInitError(connection)

    try:
        access_token, _ = refresh_oauth("airtable", connection)
    except Exception as e:
        raise OAuthRefreshError(connection, str(e)) from e

    return Api(access_token)
