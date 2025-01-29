import os

from .connections import check_connection_name
from .errors import ConnectionInitError

from auth0.management import Auth0


def auth0_client(connection: str, **kwargs) -> Auth0:
    """Initialize an Auth0 client, based on an AutoKitteh connection.

    API reference:
    https://auth0-python.readthedocs.io/en/latest/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Auth0 SDK client.

    Raises:
        ConnectionInitError: If the connection is not initialized.
        ValueError: If the connection name is invalid.
    """
    check_connection_name(connection)

    token = os.getenv(connection + "__oauth_AccessToken")
    domain = os.getenv(connection + "__auth0_domain")
    if not token or not domain:
        raise ConnectionInitError(connection)

    return Auth0(domain, token, **kwargs)
