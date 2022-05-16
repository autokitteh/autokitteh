from typing import Optional, Any

import autokitteh.proto.eventsrcsvc.svc_pb2 as eventsrcsvc_pb
import autokitteh.proto.eventsrcsvc.svc_pb2_grpc as eventsrcsvc_grpc
import autokitteh.proto.eventsvc.svc_pb2 as eventsvc_pb
import autokitteh.proto.eventsvc.svc_pb2_grpc as eventsvc_grpc

from autokitteh.api import (
    Client,
    EventID,
    EventSourceID,
    EventSourceProjectBinding,
    ProjectID,
    Value,
)


class EventSource(object):
    _id: EventSourceID
    _client: Client

    def __init__(self, client: Client, id: EventSourceID) -> None:
        self._id = id
        self._client = client

    @property
    def id(self) -> EventSourceID:
        return self._id

    def bind(
        self,
        name: Optional[str],
        project_id: ProjectID,
        assoc: str,
        config: str = '',
        approved: bool = False,
    ) -> None:
        self._client.eventsrcsvc.AddEventSourceProjectBinding(
            eventsrcsvc_pb.AddEventSourceProjectBindingRequest(
                event_source_id=self._id,
                project_id=project_id,
                name=name or '',
                association_token=assoc,
                source_config=config,
                approved=approved,
                data=None,
            ),
        )

    def get_bindings(
        self,
        project_id: Optional[ProjectID] = None,
        name: Optional[str] = None,
        assoc: Optional[str] = None,
        include_unapproved: bool = False,
    ) -> list[EventSourceProjectBinding]:
        resp = self._client.eventsrcsvc.GetEventSourceProjectBindings(
            eventsrcsvc_pb.GetEventSourceProjectBindingsRequest(
                event_source_id=self._id,
                project_id=project_id or '',
                name=name or '',
                association_token=assoc or '',
                include_unapproved=include_unapproved,
            )
        )

        return [EventSourceProjectBinding(pb) for pb in resp.bindings]

    def send(
        self,
        assoc: str,
        type_: str,
        data: dict[str, Any],
        orig_id: str,
        memo: dict[str, str]
    ) -> EventID:
        resp = self._client.eventsvc.IngestEvent(
            eventsvc_pb.IngestEventRequest(
                src_id=self._id,
                association_token=assoc,
                type=type_,
                data={k: Value.wrap(v).pb for k, v in data.items()},
                original_id=orig_id,
                memo=memo,
            ),
        )

        return EventID(resp.id)
