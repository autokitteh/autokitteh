from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
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
    __slots__ = ["name", "location", "locals"]
    class LocalsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    LOCALS_FIELD_NUMBER: _ClassVar[int]
    name: str
    location: CodeLocation
    locals: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, name: _Optional[str] = ..., location: _Optional[_Union[CodeLocation, _Mapping]] = ..., locals: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ["value", "callstack", "extra"]
    class ExtraEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUE_FIELD_NUMBER: _ClassVar[int]
    CALLSTACK_FIELD_NUMBER: _ClassVar[int]
    EXTRA_FIELD_NUMBER: _ClassVar[int]
    value: _values_pb2.Value
    callstack: _containers.RepeatedCompositeFieldContainer[CallFrame]
    extra: _containers.ScalarMap[str, str]
    def __init__(self, value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., callstack: _Optional[_Iterable[_Union[CallFrame, _Mapping]]] = ..., extra: _Optional[_Mapping[str, str]] = ...) -> None: ...
