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
    __slots__ = ["event_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    def __init__(self, event_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["event"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _event_pb2.Event
    def __init__(self, event: _Optional[_Union[_event_pb2.Event, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["integration_id", "connection_id", "event_type"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    connection_id: str
    event_type: str
    def __init__(self, integration_id: _Optional[str] = ..., connection_id: _Optional[str] = ..., event_type: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_event_pb2.Event]
    def __init__(self, events: _Optional[_Iterable[_Union[_event_pb2.Event, _Mapping]]] = ...) -> None: ...

class ListEventRecordsRequest(_message.Message):
    __slots__ = ["event_id", "state"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    state: _event_pb2.EventState
    def __init__(self, event_id: _Optional[str] = ..., state: _Optional[_Union[_event_pb2.EventState, str]] = ...) -> None: ...

class ListEventRecordsResponse(_message.Message):
    __slots__ = ["records"]
    RECORDS_FIELD_NUMBER: _ClassVar[int]
    records: _containers.RepeatedCompositeFieldContainer[_event_pb2.EventRecord]
    def __init__(self, records: _Optional[_Iterable[_Union[_event_pb2.EventRecord, _Mapping]]] = ...) -> None: ...

class AddEventRecordRequest(_message.Message):
    __slots__ = ["record"]
    RECORD_FIELD_NUMBER: _ClassVar[int]
    record: _event_pb2.EventRecord
    def __init__(self, record: _Optional[_Union[_event_pb2.EventRecord, _Mapping]] = ...) -> None: ...

class AddEventRecordResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
