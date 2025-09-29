"""Send and receive signals."""

from datetime import timedelta
from dataclasses import dataclass


@dataclass
class Signal:
    name: str
    payload: any = None


def signal(session_id: str, name: str, payload: any = None) -> None:
    """Signal a session.

    Works both for durable and non-durable sessions.
    Note that only durable sessions can receive signals.
    """
    # Dummy implementation for local development.
    pass


def next_signal(
    name: str | list[str], *, timeout: timedelta | int | float = None
) -> Signal | None:
    """Get the next signal.

    Cannot be used in an activity.
    Works only for durable sessions.
    """
    # Dummy implementation for local development.
    return None
