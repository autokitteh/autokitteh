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


def encode_jwt(payload: dict[str, int], connection: str, algorithm: str) -> str:
    """Mock function to generate JWTs, overridden by the AutoKitteh runner."""
    print("!!!!!!!!!! SDK's encode_jwt() not overridden !!!!!!!!!!")
    return "DUMMY JWT"


def refresh_oauth(integration: str, connection: str) -> tuple[str, datetime]:
    """Mock function to refresh OAuth tokens, overridden by the AutoKitteh runner."""
    print("!!!!!!!!!! SDK's refresh_oauth() not overridden !!!!!!!!!!")
    return "DUMMY TOKEN", datetime.now(UTC).replace(tzinfo=None)
