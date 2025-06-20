from typing import Any

_local_dev_store = {}


def mutate_value(key: str, op: str, *args: list[Any]) -> Any:
    """Mutate a stored value."""
    # Dummy implementation for local development.
    return {
        "set": set_value,
        "get": get_value,
        "del": del_value,
    }[op](key, *args)


def get_value(key: str) -> Any:
    """Get a stored value."""
    # Dummy implementation for local development.
    return _local_dev_store.get(key)


def set_value(key: str, value: Any) -> None:
    """Set a stored value."""
    # Dummy implementation for local development.
    _local_dev_store[key] = value


def del_value(key: str) -> None:
    """Delete a stored value."""
    # Dummy implementation for local development.
    try:
        del _local_dev_store[key]
    except KeyError:
        pass


def list_values_keys() -> list[str]:
    """List all stored keys."""
    # Dummy implementation for local development.
    return sorted(list(_local_dev_store.keys()))
