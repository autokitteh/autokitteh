"""Initialize a GitHub client, based on an AutoKitteh connection."""

import os
from urllib.parse import urljoin

from github import Auth, Github, GithubIntegration

from .connections import check_connection_name
from .errors import ConnectionInitError, EnvVarError


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
        EnvVarError: Required environment variable is missing or invalid.
    """
    check_connection_name(connection)

    # Optional: GitHub Enterprise Server
    base_url = os.getenv("GITHUB_ENTERPRISE_URL")
    if base_url:
        kwargs["base_url"] = urljoin(base_url, "api/v3")
        print("GitHub Enterprise base URL: " + kwargs["base_url"])

    # PAT + webhook
    pat = os.getenv(f"{connection}__pat")
    if pat:
        return Github(Auth.Token(pat), **kwargs)

    # GitHub App (JWT)
    private_key = os.getenv("GITHUB_PRIVATE_KEY")
    if not private_key:
        raise EnvVarError("GITHUB_PRIVATE_KEY", "missing")

    app_id = os.getenv(f"{connection}__app_id")
    if not app_id:
        raise ConnectionInitError(connection)

    install_id = os.getenv(f"{connection}__install_id")
    if not install_id:
        raise ConnectionInitError(connection)

    app = GithubIntegration(auth=Auth.AppAuth(int(app_id), private_key), **kwargs)
    return app.get_github_for_installation(int(install_id))
