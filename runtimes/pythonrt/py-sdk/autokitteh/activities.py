"""Decorator to mark a function as a Temporal activity."""

from typing import Callable

ACTIVITY_ATTR = "__ak_activity__"
INHIBIT_ACTIVITIES_ATTR = "__ak_inhibit_activities__"


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


def inhibit_activities(fn: callable) -> callable:
    """Decorator to inhibit the execution of Temporal activities.

    Functions using this decorator will not spawn activities even if otherwise
    activities should have been launched. This is useful for performing operations
    that are required to run even on replay (such as various clients creation) and
    are completely deterministic. The function results are not cached and would be
    rerun in case of replay.

    CAVEAT: Do not use this on functions that take a long time to run (more than a second),
    as they will cause the workflow to timeout.
    """
    setattr(fn, INHIBIT_ACTIVITIES_ATTR, True)
    return fn


_no_activity = set()


def register_no_activity(items: list[Callable]) -> None:
    """Mark items that should not run as activities.

    Items should be callable and hashable. If an item is a class, all methods in the
    class are marked as non-activities.

    This helps speeding up your code, but you might risk non-deterministic behavior.
    """
    _no_activity.update(items)
