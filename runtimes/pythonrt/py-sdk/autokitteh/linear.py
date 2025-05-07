"""Initialize a Linear client, based on an AutoKitteh connection."""

from requests import Session

from .oauth2_session import OAuth2Session


def linear_client(connection: str) -> Session:
    """Initialize a Linear client, based on an AutoKitteh connection.

    API reference:
    https://linear.app/developers/graphql

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Requests session object.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    return OAuth2Session("linear", connection)
