from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EventState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    EVENT_STATE_UNSPECIFIED: _ClassVar[EventState]
    EVENT_STATE_SAVED: _ClassVar[EventState]
    EVENT_STATE_PROCESSING: _ClassVar[EventState]
    EVENT_STATE_COMPLETED: _ClassVar[EventState]
    EVENT_STATE_FAILED: _ClassVar[EventState]
EVENT_STATE_UNSPECIFIED: EventState
EVENT_STATE_SAVED: EventState
EVENT_STATE_PROCESSING: EventState
EVENT_STATE_COMPLETED: EventState
EVENT_STATE_FAILED: EventState

class EventRecord(_message.Message):
    __slots__ = ["seq", "event_id", "state", "created_at"]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    seq: int
    event_id: str
    state: EventState
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, seq: _Optional[int] = ..., event_id: _Optional[str] = ..., state: _Optional[_Union[EventState, str]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ["event_id", "integration_id", "integration_token", "event_type", "data", "memo", "created_at", "seq"]
    class DataEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    class MemoEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_TOKEN_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    MEMO_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    integration_id: str
    integration_token: str
    event_type: str
    data: _containers.MessageMap[str, _values_pb2.Value]
    memo: _containers.ScalarMap[str, str]
    created_at: _timestamp_pb2.Timestamp
    seq: int
    def __init__(self, event_id: _Optional[str] = ..., integration_id: _Optional[str] = ..., integration_token: _Optional[str] = ..., event_type: _Optional[str] = ..., data: _Optional[_Mapping[str, _values_pb2.Value]] = ..., memo: _Optional[_Mapping[str, str]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., seq: _Optional[int] = ...) -> None: ...
