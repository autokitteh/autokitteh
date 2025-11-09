from collections.abc import MutableMapping
from enum import StrEnum
from typing import Any

# Dummy implementation for local development.
_local_dev_store = {}


class Store(MutableMapping):
    """Store it a dict like interface to ak store.

    Note that read-modify-write operations are not atomic.

    Values must be pickleable, see
    https://docs.python.org/3/library/pickle.html#what-can-be-pickled-and-unpickled

    Works both for durable and non-durable sessions.
    """

    def __getitem__(self, key):
        return get_value(key)

    def __setitem__(self, key, value):
        set_value(key, value)

    def __delitem__(self, key):
        del_value(key)

    def __iter__(self):
        return iter(list_values_keys())

    def __len__(self):
        return sum(1 for _ in self)


store = Store()


class Op(StrEnum):
    """Enum for operation types."""

    SET = "set"
    GET = "get"
    DEL = "del"
    ADD = "add"
    CHECK_AND_SET = "check_and_set"


def mutate_value(key: str, op: Op, *args: list[Any]) -> Any:
    """Mutate a stored value.

    Works both for durable and non-durable sessions.

    Args:
        key: Key of the value to mutate.
        op: Operation to perform on the value.
        args: Additional arguments for the operation.

    Returns:
        Any: Result of the operation, if applicable.

    Raises:
        AutoKittehError: Value is too large.
    """
    # Dummy implementation for local development.
    return {
        Op.SET: set_value,
        Op.GET: get_value,
        Op.DEL: del_value,
        Op.ADD: add_values,
        Op.CHECK_AND_SET: check_and_set_value,
    }[op](key, *args)


def get_value(key: str) -> Any:
    """Get a stored value.

    Works both for durable and non-durable sessions.

    Args:
        key: Key of the value to retrieve.

    Returns:
        Any: The stored value, or None if not found.
    """

    # Dummy implementation for local development.
    return _local_dev_store.get(key)


def check_and_set_value(key: str, expected_value: Any, new_value: Any) -> bool:
    """Check and set a stored value.

    This operation is atomic.

    Works both for durable and non-durable sessions.

    Args:
        key: Key of the value to set.
        expected_value: Expected current value.
        new_value: New value to store if the current value matches the expected value.
    Returns:
        bool: True if the value was set, False otherwise.
    """

    # Dummy implementation for local development, not atomic.
    current_value = _local_dev_store.get(key)
    if current_value == expected_value:
        _local_dev_store[key] = new_value
        return True
    return False


def set_value(key: str, value: Any) -> None:
    """Set a stored value.

    Works both for durable and non-durable sessions.

    Args:
        key: Key of the value to set.
        value: Value to store. If Value is None, it will be deleted. Value must be serializable.

    Returns:
        None.

    Raises:
        AutoKittehError: Value is too large.
    """

    # Dummy implementation for local development.
    _local_dev_store[key] = value


def add_values(key: str, value: int | float) -> int | float:
    """Add to a stored value.

    This operation is atomic.

    If key is not found, its initial value is set to the provided value.

    Works both for durable and non-durable sessions.

    Args:
        key: Key of the value to set.
        value: Value to add. Value must be serializable.

    Returns:
        New result value. Always the same type as the value stored under the key.
    """

    # Dummy implementation for local development.
    _local_dev_store[key] = _local_dev_store.get(key, 0) + value
    return _local_dev_store[key]


def del_value(key: str) -> None:
    """Delete a stored value.

    Works both for durable and non-durable sessions.

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

    Works both for durable and non-durable sessions.

    Returns:
        list[str]: Sorted list of all keys in the store.
    """

    # Dummy implementation for local development.
    return sorted(list(_local_dev_store.keys()))
