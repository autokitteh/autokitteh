from autokitteh_pb.events.v1 import event_pb2 as _event_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DispatchRequest(_message.Message):
    __slots__ = ["event", "deployment_id", "project", "wait"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    WAIT_FIELD_NUMBER: _ClassVar[int]
    event: _event_pb2.Event
    deployment_id: str
    project: str
    wait: bool
    def __init__(self, event: _Optional[_Union[_event_pb2.Event, _Mapping]] = ..., deployment_id: _Optional[str] = ..., project: _Optional[str] = ..., wait: bool = ...) -> None: ...

class DispatchResponse(_message.Message):
    __slots__ = ["event_id", "session_ids"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_IDS_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    session_ids: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, event_id: _Optional[str] = ..., session_ids: _Optional[_Iterable[str]] = ...) -> None: ...

class RedispatchRequest(_message.Message):
    __slots__ = ["event_id", "deployment_id", "wait"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    WAIT_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    deployment_id: str
    wait: bool
    def __init__(self, event_id: _Optional[str] = ..., deployment_id: _Optional[str] = ..., wait: bool = ...) -> None: ...

class RedispatchResponse(_message.Message):
    __slots__ = ["event_id", "session_ids"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_IDS_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    session_ids: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, event_id: _Optional[str] = ..., session_ids: _Optional[_Iterable[str]] = ...) -> None: ...
