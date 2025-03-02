"""Initialize a Height client, based on an AutoKitteh connection."""

from requests import Session

from .oauth2_session import OAuth2Session


def height_client(connection: str) -> Session:
    """Initialize an Height client, based on an AutoKitteh connection.

    API reference:
    https://height-api.xyz/openapi/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Requests session object.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    return OAuth2Session("height", connection)
