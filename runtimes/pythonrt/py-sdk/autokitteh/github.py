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
    app_name = os.getenv("GITHUB_APP_NAME")
    private_key = os.getenv("GITHUB_PRIVATE_KEY")
    if app_name and private_key:
        app_id = os.getenv(f"{connection}__app_id__{app_name}")
        if not app_id:
            raise ConnectionInitError(connection)

        install_id = os.getenv(f"{connection}__install_id__{app_name}")
        if not install_id:
            raise ConnectionInitError(connection)

        app = GithubIntegration(Auth.AppAuth(int(app_id), private_key), **kwargs)
        return app.get_github_for_installation(int(install_id))

    # Errors
    elif app_name:
        raise EnvVarError("GITHUB_PRIVATE_KEY", "missing")
    elif private_key:
        raise EnvVarError("GITHUB_APP_NAME", "missing")
    else:
        raise ConnectionInitError(connection)
