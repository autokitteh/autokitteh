from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Trigger(_message.Message):
    __slots__ = ["trigger_id", "name", "connection_id", "env_id", "event_type", "code_location", "filter", "data"]
    class DataEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    CODE_LOCATION_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    trigger_id: str
    name: str
    connection_id: str
    env_id: str
    event_type: str
    code_location: _program_pb2.CodeLocation
    filter: str
    data: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, trigger_id: _Optional[str] = ..., name: _Optional[str] = ..., connection_id: _Optional[str] = ..., env_id: _Optional[str] = ..., event_type: _Optional[str] = ..., code_location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., filter: _Optional[str] = ..., data: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...
