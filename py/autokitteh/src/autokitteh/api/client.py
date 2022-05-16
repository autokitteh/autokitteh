from typing import Any
import grpc

import autokitteh.proto.eventsrcsvc.svc_pb2_grpc as eventsrcsvc_grpc
import autokitteh.proto.eventsvc.svc_pb2_grpc as eventsvc_grpc


class Client(object):
    _channel: grpc.Channel

    def __init__(self, channel: grpc.Channel) -> None:
        self._channel = channel

    @staticmethod
    def insecure(*args: Any, **kwargs: Any) -> 'Client':
        return Client(grpc.insecure_channel(*args, **kwargs))

    @property
    def eventsrcsvc(self) -> eventsrcsvc_grpc.EventSourcesStub:
        return eventsrcsvc_grpc.EventSourcesStub(self._channel)

    @property
    def eventsvc(self) -> eventsvc_grpc.EventsStub:
        return eventsvc_grpc.EventsStub(self._channel)
