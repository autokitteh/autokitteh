"""AutoKitteh Python SDK."""

# Modules imported here should use only the standard library.
# We don't want people installing everything just to use autokitteh in testing

from . import errors
from .activities import activity, inhibit_activities, register_no_activity
from .attr_dict import AttrDict
from .event import Event
from .events import next_event, start, subscribe, unsubscribe
from .outcomes import outcome, http_outcome
from .signals import Signal, next_signal, signal
from .errors import AutoKittehError
from .store import (
    add_values,
    del_value,
    get_value,
    list_values_keys,
    mutate_value,
    set_value,
    store,
)
from .triggers import get_webhook_url

__all__ = [
    "AttrDict",
    "AutoKittehError",
    "errors",
    "get_webhook_url",
    "http_outcome",
    "outcome",
    "start",
    # Activities
    "activity",
    "inhibit_activities",
    "register_no_activity",
    # Events
    "Event",
    "next_event",
    "subscribe",
    "unsubscribe",
    # Signals
    "next_signal",
    "signal",
    "Signal",
    # Values
    "add_values",
    "del_value",
    "get_value",
    "list_values_keys",
    "mutate_value",
    "set_value",
    "store",
]
