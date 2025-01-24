"""Initialize Microsoft Graph SDK clients, based on AutoKitteh connections."""

from datetime import datetime, timedelta, UTC
import os

from azure.core.credentials import AccessToken
from azure.core.credentials import TokenCredential

from .connections import check_connection_name, refresh_oauth
from .errors import ConnectionInitError


def teams_client(connection: str, **kwargs):
    """Initialize a Microsoft Teams client, based on an AutoKitteh connection.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Gmail client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    raise NotImplementedError("TODO(INT-170)")


class OAuthTokenProvider(TokenCredential):
    """OAuth 2.0 token wrapper for Microsoft Graph clients."""

    def __init__(self, connection: str) -> None:
        """Initialize based on an AutoKitteh connection's environment variables.

        Args:
            connection: AutoKitteh connection name.

        Raises:
            ValueError: AutoKitteh connection name is invalid, or not using OAuth.
            ConnectionInitError: AutoKitteh connection was not initialized yet.
            OAuthRefreshError: OAuth token refresh failed.
        """
        self._connection = connection
        check_connection_name(connection)

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

    def get_token(self, *scopes: str, **kwargs) -> AccessToken:
        if self._expiry + timedelta(minutes=5) >= datetime.now(UTC):
            self._refresh_oauth_token()

        expires_on = int(self._expiry.timestamp())
        return AccessToken(token=self._access_token, expires_on=expires_on)

    def _refresh_oauth_token(self) -> None:
        self._access_token, self._expiry = refresh_oauth("microsoft", self._connection)
        self._expiry = self._expiry.astimezone(UTC)
