"""AutoKitteh SDK errors."""

from google.auth.exceptions import RefreshError


class AutoKittehError(Exception):
    """Generic base class for all errors in the AutoKitteh SDK."""

    def __init__(self, *args):
        super().__init__(*args)


class ConnectionInitError(AutoKittehError):
    """A required AutoKitteh connection was not initialized yet."""

    def __init__(self, connection: str):
        super().__init__(f"AutoKitteh connection {connection!r} not initialized")


class EnvVarError(AutoKittehError):
    """A required environment variable is missing or invalid."""

    def __init__(self, env_var: str, desc: str):
        super().__init__(f"Environment variable {env_var!r} is {desc}")


class OAuthRefreshError(AutoKittehError):
    """OAuth token refresh failed."""

    def __init__(self, connection: str, error: RefreshError):
        super().__init__(f"OAuth refresh failed for {connection!r} connection: {error}")


class AtlassianOAuthError(AutoKittehError):
    """API calls not supported by OAuth-based Atlassian connections."""

    def __init__(self, connection: str):
        msg = f"API calls not supported by {connection!r}, "
        msg += "use a token-based connection instead"
        super().__init__(msg)
