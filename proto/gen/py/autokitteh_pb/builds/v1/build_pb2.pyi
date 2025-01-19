from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class BuildStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    BUILD_STATUS_UNSPECIFIED: _ClassVar[BuildStatus]
    BUILD_STATUS_PENDING: _ClassVar[BuildStatus]
    BUILD_STATUS_IN_PROGRESS: _ClassVar[BuildStatus]
    BUILD_STATUS_READY: _ClassVar[BuildStatus]
    BUILD_STATUS_ERROR: _ClassVar[BuildStatus]
BUILD_STATUS_UNSPECIFIED: BuildStatus
BUILD_STATUS_PENDING: BuildStatus
BUILD_STATUS_IN_PROGRESS: BuildStatus
BUILD_STATUS_READY: BuildStatus
BUILD_STATUS_ERROR: BuildStatus

class Build(_message.Message):
    __slots__ = ["build_id", "project_id", "status", "error", "created_at", "updated_at"]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    project_id: str
    status: BuildStatus
    error: _program_pb2.Error
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, build_id: _Optional[str] = ..., project_id: _Optional[str] = ..., status: _Optional[_Union[BuildStatus, str]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
