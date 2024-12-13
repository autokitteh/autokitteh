import os

from .connections import check_connection_name
from .errors import ConnectionInitError

from auth0.management import Auth0


def auth0_client(connection: str, **kwargs) -> Auth0:
    """Initialize an Auth0 client, based on an AutoKitteh connection.

    API reference:
    https://auth0.com/docs/libraries/auth0-python

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Auth0 SDK client.
    """
    check_connection_name(connection)

    token = os.getenv(connection + "__oauth_AccessToken")
    print("token", token)

    # for key in os.environ:
    #     print(key, os.environ[key])

    if not token:
        raise ConnectionInitError(connection)

    # TODO: Get domain from connection.
    domain = "dev-u4mwzrvhp856wtpc.us.auth0.com"

    return Auth0(domain, token)
