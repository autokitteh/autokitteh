"""AutoKitteh Event class"""

from dataclasses import dataclass

from .attr_dict import AttrDict


@dataclass
class Event:
    """AutoKitteh Event."""

    data: AttrDict

    event_id: str | None
    """None if manual start"""

    event_type: str | None
    """None if manual start"""
