"""AutoKitteh Event class"""

from dataclasses import dataclass
from .attr_dict import AttrDict


@dataclass
class Event:
    """AutoKitteh Event as passed to entrypoints."""

    data: AttrDict
    session_id: str
