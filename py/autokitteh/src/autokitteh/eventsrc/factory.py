from typing import Optional

from os import getenv

from autokitteh.api import (
    Client,
    EventSourceID,
    AutoKittehException,
)


from .eventsrc import EventSource


_sources: dict[EventSourceID, EventSource] = {}


def _getenv(prefix: str, k: str, d: str) -> str:
    if prefix:
        prefix += '_'

    return getenv((prefix + k).upper(), d)


def init(prefix: str, id: Optional[EventSourceID] = None) -> EventSource:
    if id is None:
        src_id = _getenv(prefix, 'SRC_ID', '')
        if not src_id:
            raise AutoKittehException('id must be set')

        id = EventSourceID(src_id)

    src = _sources.get(id)
    if src:
        return src

    client = Client.insecure(target=_getenv(prefix, 'AKD_ADDR', getenv('AK_GRPC_ADDR', '127.0.0.1:50001')))

    src = EventSource(client=client, id=id)

    _sources[id] = src

    return src
