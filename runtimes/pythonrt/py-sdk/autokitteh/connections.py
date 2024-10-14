"""AutoKitteh connection-related utilities."""

from datetime import UTC, datetime
import re


def check_connection_name(connection: str) -> None:
    """Check that the given AutoKitteh connection name is valid.

    Args:
        connection: AutoKitteh connection name.

    Raises:
        ValueError: The connection name is invalid.
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError(f"Invalid AutoKitteh connection name: {connection!r}")


def refresh_oauth(integration: str, connection: str) -> tuple[str, datetime]:
    """Mock function to refresh OAuth tokens, overriden by AutoKitteh runner."""
    print("!!!!!!!!!! SDK's refresh_oauth not overriden !!!!!!!!!!")
    return "DUMMY TOKEN", datetime.now(UTC).replace(tzinfo=None)
