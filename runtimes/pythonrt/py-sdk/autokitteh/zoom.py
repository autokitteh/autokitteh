"""Initialize a Zoom client, based on an AutoKitteh connection."""

from requests import Session

from .oauth2_session import OAuth2Session


def zoom_client(connection: str) -> Session:
    """Initialize a Zoom client, based on an AutoKitteh connection.

    API reference:
    https://developers.zoom.us/docs/api/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Requests session object.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    return OAuth2Session("zoom", connection)
