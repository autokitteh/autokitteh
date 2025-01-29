"""Initialize an Atlassian client, based on an AutoKitteh connection."""

from datetime import UTC, datetime
import re
import os

from atlassian import Confluence, Jira
from requests_oauthlib import OAuth2Session

from .connections import check_connection_name
from .errors import AtlassianOAuthError, ConnectionInitError, EnvVarError


__TOKEN_URL = "https://auth.atlassian.com/oauth/token"


def jira_client(connection: str, **kwargs) -> Jira:
    """Initialize an Atlassian Jira client, based on an AutoKitteh connection.

    API reference:
    https://atlassian-python-api.readthedocs.io/jira.html

    Code samples:
    https://github.com/atlassian-api/atlassian-python-api/tree/master/examples/jira

    Args:
        connection: AutoKitteh connection name.
        **kwargs: Additional keyword arguments passed to the Jira client.
            Common options include:
            - 'url': URL of the Jira instance.
            - 'username': Username for Jira authentication.
            - 'password': Password for Jira authentication.
            - 'token': API token for Jira authentication.
            - 'verify_ssl': Boolean to verify SSL certificates.
            For a full list of accepted arguments, see:
            https://github.com/atlassian-api/atlassian-python-api/blob/master/atlassian/rest_client.py#L48

    Returns:
        Atlassian-Python-API Jira client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    check_connection_name(connection)

    if os.getenv(connection + "__oauth_AccessToken"):
        # return __atlassian_cloud_oauth2(connection, "jira", Jira, **kwargs)
        raise AtlassianOAuthError(connection)

    base_url = os.getenv(connection + "__BaseURL")
    token = os.getenv(connection + "__Token")
    if token:
        email = os.getenv(connection + "__Email")
        if not email:
            return Jira(url=base_url, token=token, **kwargs)

        return Jira(url=base_url, username=email, password=token, cloud=True, **kwargs)

    raise ConnectionInitError(connection)


def __atlassian_cloud_oauth2(connection: str, affix: str, client, **kwargs):
    """Initialize a Jira client for Atlassian Cloud using OAuth 2.0."""
    token = {
        "access_token": os.getenv(connection + "__oauth_AccessToken"),
        "token_type": os.getenv(connection + "__oauth_TokenType"),
    }

    expiry = os.getenv(connection + "__oauth_Expiry")
    if not expiry:
        raise ConnectionInitError(connection)

    client_id = os.getenv(affix.upper() + "_CLIENT_ID")
    if not client_id:
        raise EnvVarError(affix.upper() + "_CLIENT_ID", "missing")

    # Convert Go's time string (e.g. "2024-06-20 19:18:17 -0700 PDT") to
    # an ISO-8601 string that Python can parse with timezone awareness.
    timestamp = re.sub(r" [A-Z]+.*", "", expiry)
    timestamp = re.sub(r"\.\d+", "", timestamp)  # Also ignore sub-second precision.
    if datetime.fromisoformat(timestamp) <= datetime.now(UTC):
        # If the access token is expired, refresh it.
        client_secret = os.getenv(affix.upper() + "_CLIENT_SECRET")
        if not client_id:
            raise EnvVarError(affix.upper() + "_CLIENT_SECRET", "missing")

        extra = {"client_id": client_id, "client_secret": client_secret}
        oauth = OAuth2Session(client_id, auto_refresh_kwargs=extra)

        refresh = os.getenv(connection + "__oauth_RefreshToken")
        if not refresh:
            raise ConnectionInitError(connection)

        token = oauth.refresh_token(__TOKEN_URL, refresh_token=refresh)

        # Support long-running workflows - update the connection's variables.
        os.environ[connection + "__oauth_AccessToken"] = token["access_token"]
        os.environ[connection + "__oauth_RefreshToken"] = token["refresh_token"]
        expiry = datetime.fromtimestamp(token["expires_at"]).astimezone(UTC)
        os.environ[connection + "__oauth_Expiry"] = expiry.isoformat()

    cloud_id = os.getenv(connection + "__AccessID")
    if not cloud_id:
        raise ConnectionInitError(connection)

    return client(
        url=f"https://api.atlassian.com/ex/{affix.lower()}/{cloud_id}",
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
        AtlassianOAuthError
    """
    check_connection_name(connection)

    if os.getenv(connection + "__oauth_AccessToken"):
        # return __atlassian_cloud_oauth2(connection, "confluence", Confluence, **kwargs)
        raise AtlassianOAuthError(connection)

    base_url = os.getenv(connection + "__BaseURL")
    token = os.getenv(connection + "__Token")
    if token:
        email = os.getenv(connection + "__Email")
        if not email:
            return Confluence(url=base_url, token=token, **kwargs)

        return Confluence(
            url=base_url, username=email, password=token, cloud=True, **kwargs
        )

    raise ConnectionInitError(connection)


def get_base_url(connection: str) -> str | None:
    """Get the base URL of an AutoKitteh connection's Atlassian server.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Base URL of the Atlassian connection, or None if
        the AutoKitteh connection was not initialized yet.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
    """
    check_connection_name(connection)
    return os.getenv(connection + "__BaseURL") or os.getenv(connection + "__AccessURL")
