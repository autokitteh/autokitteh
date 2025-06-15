def mutate_value(key: str, op: str, *args: list[any]) -> any:
    """Mutate a stored value."""
    # Dummy implementation for local development.
    pass


def get_value(key: str) -> any:
    """Get a stored value."""
    return mutate_value(key, "get")


def set_value(key: str, value: any) -> None:
    """Set a stored value."""
    mutate_value(key, "set", value)
