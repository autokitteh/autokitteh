from typing import Any

# Dummy implementation for local development.
_local_dev_store = {}


def mutate_value(key: str, op: str, *args: list[Any]) -> Any:
    """Mutate a stored value.

    Args:
        key: Key of the value to mutate.
        op: Operation to perform on the value. Supported operations are:
            - "set": Set the value.
            - "get": Get the value.
            - "del": Delete the value.
        args: Additional arguments for the operation.

    Returns:
        Any: Result of the operation, if applicable.

    Raises:
        AutoKittehError: Value is too large.
    """

    # Dummy implementation for local development.
    return {
        "set": set_value,
        "get": get_value,
        "del": del_value,
    }[op](key, *args)


def get_value(key: str) -> Any:
    """Get a stored value.

    Args:
        key: Key of the value to retrieve.

    Returns:
        Any: The stored value, or None if not found.
    """

    # Dummy implementation for local development.
    return _local_dev_store.get(key)


def set_value(key: str, value: Any) -> None:
    """Set a stored value.

    Args:
        key: Key of the value to set.
        value: Value to store. If Value is None, it will be deleted.

    Returns:
        None.

    Raises:
        AutoKittehError: Value is too large.
    """

    # Dummy implementation for local development.
    _local_dev_store[key] = value


def del_value(key: str) -> None:
    """Delete a stored value.

    Args:
        key: Key of the value to set.

    Returns:
        None.
    """

    # Dummy implementation for local development.
    try:
        del _local_dev_store[key]
    except KeyError:
        pass


def list_values_keys() -> list[str]:
    """List all stored keys.

    Returns:
        list[str]: Sorted list of all keys in the store.
    """

    # Dummy implementation for local development.
    return sorted(list(_local_dev_store.keys()))
