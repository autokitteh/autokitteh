from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ApplyLog(_message.Message):
    __slots__ = ["message", "data"]
    class DataEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    message: str
    data: _containers.ScalarMap[str, str]
    def __init__(self, message: _Optional[str] = ..., data: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ApplyOperation(_message.Message):
    __slots__ = ["descritpion"]
    DESCRITPION_FIELD_NUMBER: _ClassVar[int]
    descritpion: str
    def __init__(self, descritpion: _Optional[str] = ...) -> None: ...

class ApplyRequest(_message.Message):
    __slots__ = ["manifest", "path"]
    MANIFEST_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    manifest: str
    path: str
    def __init__(self, manifest: _Optional[str] = ..., path: _Optional[str] = ...) -> None: ...

class ApplyResponse(_message.Message):
    __slots__ = ["logs"]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, logs: _Optional[_Iterable[str]] = ...) -> None: ...

class PlanRequest(_message.Message):
    __slots__ = ["manifest"]
    MANIFEST_FIELD_NUMBER: _ClassVar[int]
    manifest: str
    def __init__(self, manifest: _Optional[str] = ...) -> None: ...

class PlanResponse(_message.Message):
    __slots__ = ["logs"]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, logs: _Optional[_Iterable[str]] = ...) -> None: ...
