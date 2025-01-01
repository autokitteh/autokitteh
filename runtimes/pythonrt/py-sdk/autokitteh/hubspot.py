"""Initialize a HubSpot client, based on an AutoKitteh connection."""

import os

from hubspot import HubSpot

from .connections import check_connection_name, refresh_oauth
from .errors import ConnectionInitError, OAuthRefreshError


def hubspot_client(connection: str, **kwargs) -> HubSpot:
    """Initialize a HubSpot client, based on an AutoKitteh connection.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        HubSpot SDK client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    check_connection_name(connection)

    refresh_token = os.getenv(connection + "__oauth_RefreshToken")
    if not refresh_token:
        raise ConnectionInitError(connection)

    try:
        access_token, _ = refresh_oauth("hubspot", connection)
    except Exception as e:
        raise OAuthRefreshError(connection, str(e)) from e

    return HubSpot(access_token=access_token, **kwargs)
