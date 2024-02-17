from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetRequest(_message.Message):
    __slots__ = ["env_id", "project_id", "keys"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    KEYS_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    project_id: str
    keys: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, env_id: _Optional[str] = ..., project_id: _Optional[str] = ..., keys: _Optional[_Iterable[str]] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, values: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["env_id", "project_id"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    project_id: str
    def __init__(self, env_id: _Optional[str] = ..., project_id: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["keys"]
    KEYS_FIELD_NUMBER: _ClassVar[int]
    keys: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, keys: _Optional[_Iterable[str]] = ...) -> None: ...
