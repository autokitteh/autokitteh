"""Un/subscribe and consume AutoKitteh connection events."""

from datetime import timedelta
from uuid import uuid4

from .attr_dict import AttrDict


def subscribe(source: str, filter: str = "") -> str:
    """Subscribe to events on connection. Optional filter is a CEL expression."""
    # Dummy implementation for local development.
    return f"sig_{uuid4().hex}"


def unsubscribe(subscription_id: str) -> None:
    """Unsubscribe from events."""
    # Dummy implementation for local development.
    pass


def next_event(
    subscription_id: str | list[str], *, timeout: timedelta | int | float = None
) -> AttrDict:
    """Get the next event from the subscription(s).

    If timeout is not None and there are no new events after timeout, this function will
    return None.
    """
    # Dummy implementation for local development.
    return AttrDict()


def start(
    loc: str,
    data: dict | None = None,
    memo: dict | None = None,
    project: str = "",
) -> str:
    """Start a new session."""
    # Dummy implementation for local development.
    return f"ses_{uuid4().hex}"
