from autokitteh_pb.module.v1 import module_pb2 as _module_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Nothing(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class String(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: str
    def __init__(self, v: _Optional[str] = ...) -> None: ...

class Integer(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: int
    def __init__(self, v: _Optional[int] = ...) -> None: ...

class Float(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: float
    def __init__(self, v: _Optional[float] = ...) -> None: ...

class Boolean(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: bool
    def __init__(self, v: bool = ...) -> None: ...

class Symbol(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class List(_message.Message):
    __slots__ = ["vs"]
    VS_FIELD_NUMBER: _ClassVar[int]
    vs: _containers.RepeatedCompositeFieldContainer[Value]
    def __init__(self, vs: _Optional[_Iterable[_Union[Value, _Mapping]]] = ...) -> None: ...

class Set(_message.Message):
    __slots__ = ["vs"]
    VS_FIELD_NUMBER: _ClassVar[int]
    vs: _containers.RepeatedCompositeFieldContainer[Value]
    def __init__(self, vs: _Optional[_Iterable[_Union[Value, _Mapping]]] = ...) -> None: ...

class Bytes(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: bytes
    def __init__(self, v: _Optional[bytes] = ...) -> None: ...

class Dict(_message.Message):
    __slots__ = ["items"]
    class Item(_message.Message):
        __slots__ = ["k", "v"]
        K_FIELD_NUMBER: _ClassVar[int]
        V_FIELD_NUMBER: _ClassVar[int]
        k: Value
        v: Value
        def __init__(self, k: _Optional[_Union[Value, _Mapping]] = ..., v: _Optional[_Union[Value, _Mapping]] = ...) -> None: ...
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[Dict.Item]
    def __init__(self, items: _Optional[_Iterable[_Union[Dict.Item, _Mapping]]] = ...) -> None: ...

class Time(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: _timestamp_pb2.Timestamp
    def __init__(self, v: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class Duration(_message.Message):
    __slots__ = ["v"]
    V_FIELD_NUMBER: _ClassVar[int]
    v: _duration_pb2.Duration
    def __init__(self, v: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class Struct(_message.Message):
    __slots__ = ["ctor", "fields"]
    class FieldsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Value, _Mapping]] = ...) -> None: ...
    CTOR_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    ctor: Value
    fields: _containers.MessageMap[str, Value]
    def __init__(self, ctor: _Optional[_Union[Value, _Mapping]] = ..., fields: _Optional[_Mapping[str, Value]] = ...) -> None: ...

class Module(_message.Message):
    __slots__ = ["name", "members"]
    class MembersEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Value, _Mapping]] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    MEMBERS_FIELD_NUMBER: _ClassVar[int]
    name: str
    members: _containers.MessageMap[str, Value]
    def __init__(self, name: _Optional[str] = ..., members: _Optional[_Mapping[str, Value]] = ...) -> None: ...

class Function(_message.Message):
    __slots__ = ["executor_id", "name", "desc", "data", "flags"]
    EXECUTOR_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESC_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    FLAGS_FIELD_NUMBER: _ClassVar[int]
    executor_id: str
    name: str
    desc: _module_pb2.Function
    data: bytes
    flags: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, executor_id: _Optional[str] = ..., name: _Optional[str] = ..., desc: _Optional[_Union[_module_pb2.Function, _Mapping]] = ..., data: _Optional[bytes] = ..., flags: _Optional[_Iterable[str]] = ...) -> None: ...

class Value(_message.Message):
    __slots__ = ["nothing", "boolean", "string", "integer", "float", "list", "set", "dict", "bytes", "time", "duration", "struct", "module", "symbol", "function"]
    NOTHING_FIELD_NUMBER: _ClassVar[int]
    BOOLEAN_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    INTEGER_FIELD_NUMBER: _ClassVar[int]
    FLOAT_FIELD_NUMBER: _ClassVar[int]
    LIST_FIELD_NUMBER: _ClassVar[int]
    SET_FIELD_NUMBER: _ClassVar[int]
    DICT_FIELD_NUMBER: _ClassVar[int]
    BYTES_FIELD_NUMBER: _ClassVar[int]
    TIME_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    STRUCT_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    FUNCTION_FIELD_NUMBER: _ClassVar[int]
    nothing: Nothing
    boolean: Boolean
    string: String
    integer: Integer
    float: Float
    list: List
    set: Set
    dict: Dict
    bytes: Bytes
    time: Time
    duration: Duration
    struct: Struct
    module: Module
    symbol: Symbol
    function: Function
    def __init__(self, nothing: _Optional[_Union[Nothing, _Mapping]] = ..., boolean: _Optional[_Union[Boolean, _Mapping]] = ..., string: _Optional[_Union[String, _Mapping]] = ..., integer: _Optional[_Union[Integer, _Mapping]] = ..., float: _Optional[_Union[Float, _Mapping]] = ..., list: _Optional[_Union[List, _Mapping]] = ..., set: _Optional[_Union[Set, _Mapping]] = ..., dict: _Optional[_Union[Dict, _Mapping]] = ..., bytes: _Optional[_Union[Bytes, _Mapping]] = ..., time: _Optional[_Union[Time, _Mapping]] = ..., duration: _Optional[_Union[Duration, _Mapping]] = ..., struct: _Optional[_Union[Struct, _Mapping]] = ..., module: _Optional[_Union[Module, _Mapping]] = ..., symbol: _Optional[_Union[Symbol, _Mapping]] = ..., function: _Optional[_Union[Function, _Mapping]] = ...) -> None: ...
