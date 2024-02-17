from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CodeLocation(_message.Message):
    __slots__ = ["path", "row", "col", "name"]
    PATH_FIELD_NUMBER: _ClassVar[int]
    ROW_FIELD_NUMBER: _ClassVar[int]
    COL_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    path: str
    row: int
    col: int
    name: str
    def __init__(self, path: _Optional[str] = ..., row: _Optional[int] = ..., col: _Optional[int] = ..., name: _Optional[str] = ...) -> None: ...

class CallFrame(_message.Message):
    __slots__ = ["name", "location"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    name: str
    location: CodeLocation
    def __init__(self, name: _Optional[str] = ..., location: _Optional[_Union[CodeLocation, _Mapping]] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ["message", "callstack", "extra"]
    class ExtraEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    CALLSTACK_FIELD_NUMBER: _ClassVar[int]
    EXTRA_FIELD_NUMBER: _ClassVar[int]
    message: str
    callstack: _containers.RepeatedCompositeFieldContainer[CallFrame]
    extra: _containers.ScalarMap[str, str]
    def __init__(self, message: _Optional[str] = ..., callstack: _Optional[_Iterable[_Union[CallFrame, _Mapping]]] = ..., extra: _Optional[_Mapping[str, str]] = ...) -> None: ...
