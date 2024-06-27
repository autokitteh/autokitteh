"""Initialize an Atlassian client, based on an AutoKitteh connection."""

import os

from atlassian import Confluence, Jira
from jira import JIRA
from requests_oauthlib import OAuth2Session

from .connections import check_connection_name
from .errors import ConnectionInitError, EnvVarError


__TOKEN_URL = "https://auth.atlassian.com/oauth/token"


def atlassian_jira_client(connection: str, **kwargs) -> Jira:
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


def __atlassian_jira_client_cloud_oauth2(connection: str, **kwargs) -> Jira:
    """Initialize a Jira client for Atlassian Cloud using OAuth 2.0."""
    client_id = os.getenv("JIRA_CLIENT_ID")
    if not client_id:
        raise EnvVarError("JIRA_CLIENT_ID", "missing")

    client_secret = os.getenv("JIRA_CLIENT_SECRET")
    if not client_id:
        raise EnvVarError("JIRA_CLIENT_SECRET", "missing")

    extra = {"client_id": client_id, "client_secret": client_secret}
    oauth = OAuth2Session(client_id, auto_refresh_kwargs=extra)

    refresh = os.getenv(connection + "__oauth_RefreshToken")
    if not refresh:
        raise ConnectionInitError(connection)

    token = oauth.refresh_token(__TOKEN_URL, refresh_token=refresh)

    cloud_id = os.getenv(connection + "__AccessID")
    if not cloud_id:
        raise ConnectionInitError(connection)

    return Jira(
        url="https://api.atlassian.com/ex/jira/" + cloud_id,
        oauth2={"client_id": client_id, "token": token},
        **kwargs,
    )


def confluence_client(connection: str, **kwargs) -> Confluence:
    """Initialize an Atlassian Confluence client, based on an AutoKitteh connection.

    API reference:
    https://atlassian-python-api.readthedocs.io/confluence.html

    Code samples:
    https://github.com/atlassian-api/atlassian-python-api/tree/master/examples/confluence

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Atlassian-Python-API Confluence client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    check_connection_name(connection)

    if os.getenv(connection + "__oauth_AccessToken"):
        return __confluence_client_cloud_oauth2(connection, **kwargs)

    base_url = os.getenv(connection + "__BaseURL")
    token = os.getenv(connection + "__Token")
    if token:
        email = os.getenv(connection + "__Email")
        if not email:
            return Confluence(url=base_url, token=token, **kwargs)
        return Confluence(
            url=base_url,
            username=email,
            password=token,
            cloud=True,
            **kwargs,
        )

    raise ConnectionInitError(connection)


def __confluence_client_cloud_oauth2(connection: str, **kwargs) -> Confluence:
    """Initialize a Confluence client for Atlassian Cloud using OAuth 2.0."""
    client_id = os.getenv("CONFLUENCE_CLIENT_ID")
    if not client_id:
        raise EnvVarError("CONFLUENCE_CLIENT_ID", "missing")

    client_secret = os.getenv("CONFLUENCE_CLIENT_SECRET")
    if not client_id:
        raise EnvVarError("CONFLUENCE_CLIENT_SECRET", "missing")

    extra = {"client_id": client_id, "client_secret": client_secret}
    oauth = OAuth2Session(client_id, auto_refresh_kwargs=extra)

    refresh = os.getenv(connection + "__oauth_RefreshToken")
    if not refresh:
        raise ConnectionInitError(connection)

    token = oauth.refresh_token(__TOKEN_URL, refresh_token=refresh)

    cloud_id = os.getenv(connection + "__AccessID")
    if not cloud_id:
        raise ConnectionInitError(connection)

    return Confluence(
        url="https://api.atlassian.com/ex/confluence/" + cloud_id,
        oauth2={"client_id": client_id, "token": token},
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
