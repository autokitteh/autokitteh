"""Initialize an Atlassian Jira client, based on an AutoKitteh connection."""

from datetime import UTC, datetime
import re
import os

from atlassian import Jira
from jira import JIRA

from .connections import check_connection_name
from .errors import ConnectionInitError, EnvVarError


def atlassian_jira_client(connection: str, **kwargs):
    """Initialize an Atlassian Jira client, based on an AutoKitteh connection.

    API reference:
    https://atlassian-python-api.readthedocs.io/jira.html

    Code samples:
    https://github.com/atlassian-api/atlassian-python-api/tree/master/examples/jira

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Atlassian-Python-API Jira client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        RuntimeError: OAuth 2.0 access token expired.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    check_connection_name(connection)

    if os.getenv(connection + "__oauth_AccessToken"):
        return __atlassian_jira_client_cloud_oauth2(connection, **kwargs)

    base_url = os.getenv(connection + "__BaseURL")
    token = os.getenv(connection + "__Token")
    if token:
        email = os.getenv(connection + "__Email")
        if not email:
            return Jira(url=base_url, token=token, **kwargs)
        return Jira(
            url=base_url,
            username=email,
            password=token,
            cloud=True,
            **kwargs,
        )

    raise ConnectionInitError(connection)


def __atlassian_jira_client_cloud_oauth2(connection: str, **kwargs):
    """Initialize a Jira client for Atlassian Cloud using OAuth 2.0."""
    expiry = os.getenv(connection + "__oauth_Expiry")
    if not expiry:
        raise ConnectionInitError(connection)

    # Convert Go's time string (e.g. "2024-06-20 19:18:17 +0700 PDT") to
    # an ISO-8601 string that Python can parse with timezone awareness.
    timestamp = re.sub(r"[ A-Z]+.*", "", expiry)
    if datetime.fromisoformat(timestamp) < datetime.now(UTC):
        raise RuntimeError("OAuth 2.0 access token expired on: " + expiry)

    cloud_id = os.getenv(connection + "__access_id")
    if not cloud_id:
        raise ConnectionInitError(connection)

    client_id = os.getenv("JIRA_CLIENT_ID")
    if not client_id:
        raise EnvVarError("JIRA_CLIENT_ID", "missing")

    return Jira(
        url="https://api.atlassian.com/ex/jira/" + cloud_id,
        oauth2={
            "client_id": client_id,
            "token": {
                "access_token": os.getenv(connection + "__oauth_AccessToken"),
                "token_type": os.getenv(connection + "__oauth_TokenType"),
            },
        },
        **kwargs,
    )


def jira_client(connection: str, **kwargs):
    """Initialize an Atlassian Jira client, based on an AutoKitteh connection.

    API reference:
    https://jira.readthedocs.io/

    Code samples:
    https://github.com/pycontribs/jira/tree/main/examples

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Jira client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    check_connection_name(connection)

    base_url = os.getenv(connection + "__BaseURL")
    token = os.getenv(connection + "__Token")
    if token:
        email = os.getenv(connection + "__Email")
        if email:
            return JIRA(base_url, basic_auth=(email, token), **kwargs)
        else:
            return JIRA(base_url, token_auth=token, **kwargs)

    raise ConnectionInitError(connection)
