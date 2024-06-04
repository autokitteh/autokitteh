"""AutoKitteh Python SDK"""

from uuid import uuid4


ACTIVITY_ATTR = '__activity__'

class AttrDict(dict):
    """Allow attribute access to dictionary keys.

    >>> config = AttrDict({'server': {'port': 8080}, 'debug': True})
    >>> config.server.port
    8080
    >>> config.debug
    True
    """
    def __getattr__(self, name):
        try:
            value = self[name]
            if isinstance(value, dict):
                value = AttrDict(value)
            return value
        except KeyError:
            raise AttributeError(name)

    def __setattr__(self, attr, value):
        # The default __getattr__ doesn't fail but also don't change values
        cls = self.__class__.__name__
        raise NotImplementedError(f'{cls} does not support setting attributes')


def activity(fn: callable) -> callable:
    """Decorator to mark a function as an activity."""
    setattr(fn, ACTIVITY_ATTR, True)
    return fn


def subscribe(connection_name: str, filter: str ='') -> str:
    """Subscribe to events on connection. Option filter is a CEL expression."""

    # Dummy implementation for local development
    return uuid4().hex


def unsubscribe(id: str) -> None:
    """Unsubscribe from events."""

    # Dummy implementation for local development
    pass


def next_event(id: str) -> AttrDict:
    """Get the next event from the subscription."""

    # Dummy implementation for local development
    return AttrDict()
