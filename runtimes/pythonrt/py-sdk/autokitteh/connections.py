"""AutoKitteh connection-related utilities."""

import re
from datetime import UTC, datetime


def check_connection_name(connection: str) -> None:
    """Check that the given AutoKitteh connection name is valid.

    Args:
        connection: AutoKitteh connection name.

    Raises:
        ValueError: The connection name is invalid.
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError(f'Invalid AutoKitteh connection name: "{connection}"')


def refresh_oauth(token: str) -> tuple[str, datetime]:
    """Mock function to refresh oauth tokens, hijacked by AK runner."""
    return "", datetime.now(UTC)
