"""Initialize a Linear client, based on an AutoKitteh connection."""

import os
from requests import Session

from .connections import check_connection_name
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
    check_connection_name(connection)
    api_key = os.getenv(f"{connection}__api_key")
    if api_key:
        session = Session()
        session.headers.update(
            {"Authorization": api_key, "Content-Type": "application/json"}
        )
        session.base_url = "https://api.linear.app/graphql/"
        return session

    return OAuth2Session(
        "linear", connection, base_url="https://api.linear.app/graphql/"
    )
