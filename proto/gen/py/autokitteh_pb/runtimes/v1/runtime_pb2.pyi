from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Runtime(_message.Message):
    __slots__ = ["name", "file_extensions"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    FILE_EXTENSIONS_FIELD_NUMBER: _ClassVar[int]
    name: str
    file_extensions: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, name: _Optional[str] = ..., file_extensions: _Optional[_Iterable[str]] = ...) -> None: ...
