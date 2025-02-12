import requests
from os import getenv

from .connections import refresh_oauth
from .connections import check_connection_name
from .errors import ConnectionInitError


class OAuth2Session(requests.Session):
    """Encapsulates arequests session, based on an AutoKitteh connection.

    - Automatically sets the Authorization header with an OAuth token.
    - Automatically refreshes an OAuth token if a refresh token is
      initialized in the connection.
    """

    def __init__(self, integration: str, connection: str, *args, **kwargs):
        """Initialize a requests session, based on an AutoKitteh connection.

        Args:
            connection: AutoKitteh integration and connection name.

        Returns:
            OAuth2Session.

        Raises:
            ValueError: AutoKitteh connection name is invalid.
            ConnectionInitError: AutoKitteh connection was not initialized yet.
        """
        check_connection_name(connection)

        super().__init__(*args, **kwargs)

        self._integration = integration
        self._connection = connection
        self._refresh_token = getenv(f"{connection}__oauth_refresh_token")
        self._token = getenv(f"{connection}__oauth_access_token")

        if not self._token:
            raise ConnectionInitError(connection)

        self._set_token(self._token)

        self.hooks["response"].append(self._intercept)

    def _intercept(self, r, *args, **kwargs):
        if not self._refresh_token or r.status_code != 401:
            return r

        token, _ = refresh_oauth(self._integration, self._connection)
        self._set_token(token)
        r.request.headers["Authorization"] = f"Bearer {token}"

        return self.send(r.request)

    def _set_token(self, token):
        self._token = token
        self.headers.update({"Authorization": f"Bearer {token}"})
