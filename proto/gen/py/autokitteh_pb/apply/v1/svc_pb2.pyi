from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

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
