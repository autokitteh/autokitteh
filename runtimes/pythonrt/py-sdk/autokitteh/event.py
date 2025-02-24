"""AutoKitteh Event class"""

from dataclasses import dataclass


@dataclass
class Event:
    data: dict[str, any]
    session_id: str
