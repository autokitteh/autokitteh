from typing import NamedTuple


import autokitteh.proto.eventsrc.src_pb2 as eventsrc


class EventSourceProjectBinding(NamedTuple):
    pb: eventsrc.EventSourceProjectBinding
