from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Build(_message.Message):
    __slots__ = ["build_id", "project_id", "created_at", "status"]
    class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        STATUS_UNSPECIFIED: _ClassVar[Build.Status]
        STATUS_PENDING: _ClassVar[Build.Status]
        STATUS_RUNNING: _ClassVar[Build.Status]
        STATUS_SUCCESS: _ClassVar[Build.Status]
        STATUS_FAILED: _ClassVar[Build.Status]
    STATUS_UNSPECIFIED: Build.Status
    STATUS_PENDING: Build.Status
    STATUS_RUNNING: Build.Status
    STATUS_SUCCESS: Build.Status
    STATUS_FAILED: Build.Status
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    project_id: str
    created_at: _timestamp_pb2.Timestamp
    status: Build.Status
    def __init__(self, build_id: _Optional[str] = ..., project_id: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., status: _Optional[_Union[Build.Status, str]] = ...) -> None: ...
