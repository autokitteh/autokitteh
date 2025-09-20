from typing import Any


def outcome(v: Any) -> None:
    """Log an outcome for the current session.

    Works both in durable and nondurable sessions.

    Args:
        v: The outcome value. Can be any JSON-serializable value.

    """
    # Dummy implementation for local development.
    pass


def http_outcome(
    status_code: int = 200,
    body: Any = None,
    json: Any = None,
    headers: dict[str, str] = {},
    more: bool = False,
) -> None:
    """Respond to an HTTP request.

    Works both in durable and nondurable sessions.

    Args:
        status_code: HTTP status code to return. Ignored if not the first response to a request.
        body: body to return. If it is a dict or a list, it will be serialized as
            JSON. If it is a string or bytes, it will be returned as-is.
        json: JSON-serializable value to return as JSON. If specified, the Content-Type
            header will be set to application/json. Cannot be used together with body.
        headers: dict of headers to return.
        more: If True, indicates that more responses will follow for this request.
    """
    # Dummy implementation for local development.
    pass
