from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Trigger(_message.Message):
    __slots__ = ["trigger_id", "connection_id", "env_id", "event_type", "code_location", "filter"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    CODE_LOCATION_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    trigger_id: str
    connection_id: str
    env_id: str
    event_type: str
    code_location: _program_pb2.CodeLocation
    filter: str
    def __init__(self, trigger_id: _Optional[str] = ..., connection_id: _Optional[str] = ..., env_id: _Optional[str] = ..., event_type: _Optional[str] = ..., code_location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., filter: _Optional[str] = ...) -> None: ...
