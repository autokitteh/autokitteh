"""AutoKitteh SDK errors."""


class AutoKittehError(Exception):
    """Generic base class for all errors in the AutoKitteh SDK."""

    def __init__(self, *args):
        super().__init__(*args)


class ConnectionInitError(AutoKittehError):
    """A required AutoKitteh connection was not initialized yet."""

    def __init__(self, connection: str):
        super().__init__(f'AutoKitteh connection "{connection}" not initialized')


class EnvVarError(AutoKittehError):
    """A required environment variable is missing or invalid."""

    def __init__(self, env_var: str, desc: str):
        super().__init__(f'Environment variable "{env_var}" is {desc}')
