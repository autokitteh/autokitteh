from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DecideRequest(_message.Message):
    __slots__ = ["path", "user_id", "subject_id", "data"]
    class DataEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    PATH_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_ID_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    path: str
    user_id: str
    subject_id: str
    data: _containers.ScalarMap[str, str]
    def __init__(self, path: _Optional[str] = ..., user_id: _Optional[str] = ..., subject_id: _Optional[str] = ..., data: _Optional[_Mapping[str, str]] = ...) -> None: ...

class DecideResponse(_message.Message):
    __slots__ = ["result"]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    result: _struct_pb2.Value
    def __init__(self, result: _Optional[_Union[_struct_pb2.Value, _Mapping]] = ...) -> None: ...
