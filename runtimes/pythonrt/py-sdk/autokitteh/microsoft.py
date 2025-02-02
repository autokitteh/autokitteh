"""Initialize Microsoft Graph SDK clients, based on AutoKitteh connections."""

from datetime import datetime, timedelta, UTC
import os

from azure.core import credentials
from azure import identity
from msgraph import GraphServiceClient

from .connections import check_connection_name, refresh_oauth
from .errors import ConnectionInitError


# Default buffer time to refresh OAuth tokens before they expire.
DEFAULT_REFRESH_BUFFER_TIME = timedelta(minutes=5)


def teams_client(connection: str, **kwargs) -> GraphServiceClient:
    """Initialize a Microsoft Teams client, based on an AutoKitteh connection.

    API documentation:
    https://docs.autokitteh.com/integrations/microsoft/teams/python

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Microsoft Graph client.

    Raises:
        ValueError: AutoKitteh connection name or auth type are invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    # TODO: Narrower scopes instead of all the default ones.
    return _microsoft_client(connection, **kwargs)


def _microsoft_client(connection: str, **kwargs) -> GraphServiceClient:
    """Initialize a Microsoft Graph client, based on an AutoKitteh connection.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Microsoft Graph client.

    Raises:
        ValueError: AutoKitteh connection name or auth type are invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    check_connection_name(connection)

    auth_type = os.getenv(connection + "__auth_type", "")
    match auth_type:
        case "oauthDefault" | "oauthPrivate":
            creds = OAuthTokenProvider(connection)
        case "daemonApp":
            if os.getenv(connection + "__private_certificate"):
                creds = _certificate_credentials(connection)
            else:
                creds = _client_secret_credentials(connection)
        case _:
            err = f"AutoKitteh connection {connection!r} with "
            raise ValueError(err + f" invalid auth type {auth_type!r}")

    return GraphServiceClient(credentials=creds, **kwargs)


class OAuthTokenProvider(credentials.TokenCredential):
    """OAuth 2.0 token wrapper for Microsoft Graph clients."""

    def __init__(self, connection: str, buffer_time: timedelta | None = None) -> None:
        """Initialize credentials based on an AutoKitteh connection's variables.

        Used by server-default and private OAuth 2.0 user-delegated apps.

        Args:
            connection: AutoKitteh connection name.
            buffer_time: Buffer time to refresh OAuth tokens before they expire
                (optional, default = 5 minutes).

        Raises:
            ValueError: AutoKitteh connection name is invalid, or not using OAuth.
            ConnectionInitError: AutoKitteh connection was not initialized yet.
            OAuthRefreshError: OAuth token refresh failed.
        """
        self._connection = connection
        check_connection_name(connection)
        self._buffer_time = buffer_time or DEFAULT_REFRESH_BUFFER_TIME

        auth_type = os.getenv(connection + "__auth_type")
        if not auth_type:
            raise ConnectionInitError(connection)
        if not auth_type.startswith("oauth"):
            raise ValueError(f"AutoKitteh connection {connection!r} not using OAuth")

        self._access_token = os.getenv(connection + "__oauth_access_token", "")
        if not self._access_token:
            raise ConnectionInitError(connection)

        expiry = os.getenv(connection + "__oauth_expiry", "")
        self._expiry = datetime.fromisoformat(expiry).astimezone(UTC)
        self._refresh_token = os.getenv(connection + "__oauth_refresh_token", "")

    def get_token(self, *scopes: str, **kwargs) -> credentials.AccessToken:
        if self._expiry + self._buffer_time >= datetime.now(UTC):
            self._refresh_oauth_token()

        expires_on = int(self._expiry.timestamp())
        return credentials.AccessToken(token=self._access_token, expires_on=expires_on)

    def _refresh_oauth_token(self) -> None:
        self._access_token, self._expiry = refresh_oauth("microsoft", self._connection)
        self._expiry = self._expiry.astimezone(UTC)


def _certificate_credentials(connection: str) -> identity.CertificateCredential:
    """Initialize credentials based on an AutoKitteh connection's variables.

    Used by daemon (i.e. non-user-delegated) applications with advanced auth needs.

    Args:
        connection: AutoKitteh connection name.

    Raises:
        ValueError: AutoKitteh connection was configured to use OAuth
            (i.e. user-delegated permissions instead of application ones).
        ConnectionInitError: AutoKitteh connection was not initialized yet,
            or initialized to use a client secret instead of a certificate.
    """
    check_connection_name(connection)
    auth_type = os.getenv(connection + "__auth_type")
    if not auth_type:
        raise ConnectionInitError(connection)
    if auth_type.startswith("oauth"):
        raise ValueError(f"AutoKitteh connection {connection!r} using OAuth")

    client_id = os.getenv(connection + "__private_client_id")
    certificate = os.getenv(connection + "__private_certificate")
    if not client_id or not certificate:
        raise ConnectionInitError(connection)

    return identity.CertificateCredential(
        client_id=client_id,
        certificate_data=certificate.encode("utf-8"),
        tenant_id=os.getenv(connection + "__private_tenant_id", "") or "common",
    )


def _client_secret_credentials(connection: str) -> identity.ClientSecretCredential:
    """Initialize credentials based on an AutoKitteh connection's variables.

    Used by daemon (i.e. non-user-delegated) applications with simple auth needs.

    Args:
        connection: AutoKitteh connection name.

    Raises:
        ValueError: AutoKitteh connection was configured to use OAuth
            (i.e. user-delegated permissions instead of application ones).
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)
    auth_type = os.getenv(connection + "__auth_type")
    if not auth_type:
        raise ConnectionInitError(connection)
    if auth_type.startswith("oauth"):
        raise ValueError(f"AutoKitteh connection {connection!r} using OAuth")

    client_id = os.getenv(connection + "__private_client_id")
    client_secret = os.getenv(connection + "__private_client_secret")
    if not client_id or not client_secret:
        raise ConnectionInitError(connection)

    return identity.ClientSecretCredential(
        client_id=client_id,
        client_secret=client_secret,
        tenant_id=os.getenv(connection + "__private_tenant_id", "") or "common",
    )
