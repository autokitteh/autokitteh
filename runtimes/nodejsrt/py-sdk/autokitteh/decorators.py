"""Decorator to mark a function as a Temporal activity."""

ACTIVITY_ATTR = "__activity__"


def activity(fn: callable) -> callable:
    """Decorator to mark a function as a Temporal activity.

    This forces AutoKitteh to run the function as a single Temporal activity,
    instead of a sequence of activities within a Temporal workflow.

    Use this decorator when you want to run functions with input arguments
    and/or return values that are not compatible with pickle.

    For more details, see:
    https://docs.python.org/3/library/pickle.html#what-can-be-pickled-and-unpickled
    """
    setattr(fn, ACTIVITY_ATTR, True)
    return fn
