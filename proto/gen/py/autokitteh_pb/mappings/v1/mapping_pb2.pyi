from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class MappingEvent(_message.Message):
    __slots__ = ["event_type", "code_location"]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    CODE_LOCATION_FIELD_NUMBER: _ClassVar[int]
    event_type: str
    code_location: _program_pb2.CodeLocation
    def __init__(self, event_type: _Optional[str] = ..., code_location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ...) -> None: ...

class Mapping(_message.Message):
    __slots__ = ["mapping_id", "env_id", "connection_id", "module_name", "events"]
    MAPPING_ID_FIELD_NUMBER: _ClassVar[int]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    MODULE_NAME_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    mapping_id: str
    env_id: str
    connection_id: str
    module_name: str
    events: _containers.RepeatedCompositeFieldContainer[MappingEvent]
    def __init__(self, mapping_id: _Optional[str] = ..., env_id: _Optional[str] = ..., connection_id: _Optional[str] = ..., module_name: _Optional[str] = ..., events: _Optional[_Iterable[_Union[MappingEvent, _Mapping]]] = ...) -> None: ...
