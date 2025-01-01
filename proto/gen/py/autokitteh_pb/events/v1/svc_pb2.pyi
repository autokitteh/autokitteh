from autokitteh_pb.events.v1 import event_pb2 as _event_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SaveRequest(_message.Message):
    __slots__ = ["event"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _event_pb2.Event
    def __init__(self, event: _Optional[_Union[_event_pb2.Event, _Mapping]] = ...) -> None: ...

class SaveResponse(_message.Message):
    __slots__ = ["event_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    def __init__(self, event_id: _Optional[str] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["event_id", "json_values"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    JSON_VALUES_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    json_values: bool
    def __init__(self, event_id: _Optional[str] = ..., json_values: bool = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["event"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _event_pb2.Event
    def __init__(self, event: _Optional[_Union[_event_pb2.Event, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["integration_id", "destination_id", "event_type", "max_results", "order", "project_id", "org_id", "json_values"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    DESTINATION_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    MAX_RESULTS_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    JSON_VALUES_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    destination_id: str
    event_type: str
    max_results: int
    order: str
    project_id: str
    org_id: str
    json_values: bool
    def __init__(self, integration_id: _Optional[str] = ..., destination_id: _Optional[str] = ..., event_type: _Optional[str] = ..., max_results: _Optional[int] = ..., order: _Optional[str] = ..., project_id: _Optional[str] = ..., org_id: _Optional[str] = ..., json_values: bool = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_event_pb2.Event]
    def __init__(self, events: _Optional[_Iterable[_Union[_event_pb2.Event, _Mapping]]] = ...) -> None: ...
