"""Initialize a GitHub client, based on an AutoKitteh connection."""

import os
import time
from urllib.parse import urljoin

from github import Auth, Consts, Github, GithubIntegration

from .connections import check_connection_name, encode_jwt
from .errors import ConnectionInitError


def github_client(connection: str, **kwargs) -> Github:
    """Initialize a GitHub client, based on an AutoKitteh connection.

    API reference and examples: https://pygithub.readthedocs.io/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        PyGithub client.

    Raises:
        ValueError: AutoKitteh connection name or GitHub app IDs are invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    # Optional: GitHub Enterprise Server
    base_url = os.getenv(f"{connection}__enterprise_url") or os.getenv(
        "GITHUB_ENTERPRISE_URL"
    )

    if base_url:
        kwargs["base_url"] = urljoin(base_url, "api/v3")
        print("GitHub Enterprise base URL: " + kwargs["base_url"])

    # PAT + webhook
    pat = os.getenv(f"{connection}__pat")
    if pat:
        return Github(auth=Auth.Token(pat), **kwargs)

    # GitHub App (JWT)
    app_id = os.getenv(f"{connection}__app_id")
    if not app_id:
        raise ConnectionInitError(connection)

    install_id = os.getenv(f"{connection}__install_id")
    if not install_id:
        raise ConnectionInitError(connection)

    app = None
    private_key = os.getenv(f"{connection}__private_key")
    if private_key:
        # Exposing the private key is fine here as it belongs to the user.
        app = GithubIntegration(auth=Auth.AppAuth(int(app_id), private_key), **kwargs)
    else:
        app = GithubIntegration(auth=AppAuth(int(app_id), connection), **kwargs)
    return app.get_github_for_installation(int(install_id))


class AppAuth(Auth.AppAuth):
    """Generate JWTs without exposing the GitHub app's private key.

    Based on: https://github.com/PyGithub/PyGithub/blob/main/github/Auth.py
    """

    def __init__(self, app_id: int, ak_connection_name: str):
        self._app_id = app_id
        self._conn_name = ak_connection_name
        self._jwt_expiry = Consts.DEFAULT_JWT_EXPIRY
        self._jwt_issued_at = Consts.DEFAULT_JWT_ISSUED_AT
        self._jwt_algorithm = Consts.DEFAULT_JWT_ALGORITHM

    def create_jwt(self, expiration: int | None = None) -> str:
        now = int(time.time())
        payload = {
            "iat": now + self._jwt_issued_at,
            "exp": now + (expiration if expiration is not None else self._jwt_expiry),
            "iss": self._app_id,
        }
        # This is the only change from the original code: replace the call to jwt.encode().
        # We don't monkey-patch it because the jwt module is usable outside the GitHub client too.
        return encode_jwt(payload, self._conn_name, self._jwt_algorithm)
