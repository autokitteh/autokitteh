"""AutoKitteh Python SDK"""


ACTIVITY_ATTR = '__activity__'


def activity(fn):
    """Decorator to mark a function as an activity."""
    setattr(fn, ACTIVITY_ATTR, True)
    return fn

