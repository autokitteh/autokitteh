"""Un/subscribe and consume AutoKitteh connection events."""

from datetime import timedelta
from uuid import uuid4
from typing import Any

from .attr_dict import AttrDict


def subscribe(source: str, filter: str = "") -> str:
    """Subscribe to events on connection. Optional filter is a CEL expression.

    Cannot be used in an activity."""
    # Dummy implementation for local development.
    return f"sig_{uuid4().hex}"


def unsubscribe(subscription_id: str) -> None:
    """Unsubscribe from events.

    Cannot be used in an activity."""
    # Dummy implementation for local development.
    pass


def next_event(
    subscription_id: str | list[str], *, timeout: timedelta | int | float = None
) -> AttrDict:
    """Get the next event from the subscription(s).

    If timeout is not None and there are no new events after timeout, this function will
    return None.

    Cannot be used in an activity.
    """
    # Dummy implementation for local development.
    return AttrDict()


def start(
    loc: str,
    data: dict | None = None,
    memo: dict | None = None,
    project: str = "",
) -> str:
    """Start a new session.

    Cannot be used in an activity."""
    # Dummy implementation for local development.
    return f"ses_{uuid4().hex}"


def http_respond(
    status_code: int, body: Any = None, headers: dict | None = None, more: bool = False
) -> None:
    """Respond to an HTTP request.

    This function has an effect only within a session that was triggered by a Webhook trigger with sync_webhook=True.

    Args:
        status_code: HTTP status code to return. Ignored if not the first response to a request.
        body: body to return. If it is a dict or a list, it will be serialized as
            JSON. If it is a string or bytes, it will be returned as-is.
        headers: dict of headers to return.
        more: If True, indicates that more responses will follow for this request.
    """
    # Dummy implementation for local development.
    pass
