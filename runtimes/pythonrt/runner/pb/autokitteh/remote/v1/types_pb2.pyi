from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Frame(_message.Message):
    __slots__ = ["filename", "lineno", "code", "name"]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    LINENO_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    filename: str
    lineno: int
    code: str
    name: str
    def __init__(self, filename: _Optional[str] = ..., lineno: _Optional[int] = ..., code: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ["data"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    def __init__(self, data: _Optional[bytes] = ...) -> None: ...
