from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Trigger(_message.Message):
    __slots__ = ["trigger_id", "name", "source_type", "env_id", "event_type", "code_location", "filter", "connection_id", "schedule", "webhook_slug"]
    class SourceType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        SOURCE_TYPE_UNSPECIFIED: _ClassVar[Trigger.SourceType]
        SOURCE_TYPE_CONNECTION: _ClassVar[Trigger.SourceType]
        SOURCE_TYPE_WEBHOOK: _ClassVar[Trigger.SourceType]
        SOURCE_TYPE_SCHEDULE: _ClassVar[Trigger.SourceType]
    SOURCE_TYPE_UNSPECIFIED: Trigger.SourceType
    SOURCE_TYPE_CONNECTION: Trigger.SourceType
    SOURCE_TYPE_WEBHOOK: Trigger.SourceType
    SOURCE_TYPE_SCHEDULE: Trigger.SourceType
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    CODE_LOCATION_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    WEBHOOK_SLUG_FIELD_NUMBER: _ClassVar[int]
    trigger_id: str
    name: str
    source_type: Trigger.SourceType
    env_id: str
    event_type: str
    code_location: _program_pb2.CodeLocation
    filter: str
    connection_id: str
    schedule: str
    webhook_slug: str
    def __init__(self, trigger_id: _Optional[str] = ..., name: _Optional[str] = ..., source_type: _Optional[_Union[Trigger.SourceType, str]] = ..., env_id: _Optional[str] = ..., event_type: _Optional[str] = ..., code_location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., filter: _Optional[str] = ..., connection_id: _Optional[str] = ..., schedule: _Optional[str] = ..., webhook_slug: _Optional[str] = ...) -> None: ...
