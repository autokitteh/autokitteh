from autokitteh_pb.integrations.v1 import integration_pb2 as _integration_pb2
from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetRequest(_message.Message):
    __slots__ = ["integration_id", "name"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    name: str
    def __init__(self, integration_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["integration"]
    INTEGRATION_FIELD_NUMBER: _ClassVar[int]
    integration: _integration_pb2.Integration
    def __init__(self, integration: _Optional[_Union[_integration_pb2.Integration, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["name_substring"]
    NAME_SUBSTRING_FIELD_NUMBER: _ClassVar[int]
    name_substring: str
    def __init__(self, name_substring: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["integrations"]
    INTEGRATIONS_FIELD_NUMBER: _ClassVar[int]
    integrations: _containers.RepeatedCompositeFieldContainer[_integration_pb2.Integration]
    def __init__(self, integrations: _Optional[_Iterable[_Union[_integration_pb2.Integration, _Mapping]]] = ...) -> None: ...

class CallRequest(_message.Message):
    __slots__ = ["integration_id", "function", "args", "kwargs"]
    class KwargsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    FUNCTION_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    KWARGS_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    function: _values_pb2.Value
    args: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
    kwargs: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, integration_id: _Optional[str] = ..., function: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., args: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., kwargs: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class CallResponse(_message.Message):
    __slots__ = ["value", "error"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    value: _values_pb2.Value
    error: _program_pb2.Error
    def __init__(self, value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...

class ConfigureRequest(_message.Message):
    __slots__ = ["integration_id", "connection_id"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    connection_id: str
    def __init__(self, integration_id: _Optional[str] = ..., connection_id: _Optional[str] = ...) -> None: ...

class ConfigureResponse(_message.Message):
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

class DisconnectRequest(_message.Message):
    __slots__ = ["executor_id"]
    EXECUTOR_ID_FIELD_NUMBER: _ClassVar[int]
    executor_id: str
    def __init__(self, executor_id: _Optional[str] = ...) -> None: ...

class DisconnectResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
