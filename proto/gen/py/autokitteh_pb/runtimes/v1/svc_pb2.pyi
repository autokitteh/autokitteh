from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.runtimes.v1 import build_pb2 as _build_pb2
from autokitteh_pb.runtimes.v1 import run_pb2 as _run_pb2
from autokitteh_pb.runtimes.v1 import runtime_pb2 as _runtime_pb2
from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DescribeRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class DescribeResponse(_message.Message):
    __slots__ = ["runtime"]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    runtime: _runtime_pb2.Runtime
    def __init__(self, runtime: _Optional[_Union[_runtime_pb2.Runtime, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["runtimes"]
    RUNTIMES_FIELD_NUMBER: _ClassVar[int]
    runtimes: _containers.RepeatedCompositeFieldContainer[_runtime_pb2.Runtime]
    def __init__(self, runtimes: _Optional[_Iterable[_Union[_runtime_pb2.Runtime, _Mapping]]] = ...) -> None: ...

class BuildRequest(_message.Message):
    __slots__ = ["runtime_name", "value_names", "root_url", "path"]
    RUNTIME_NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_NAMES_FIELD_NUMBER: _ClassVar[int]
    ROOT_URL_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    runtime_name: str
    value_names: _containers.RepeatedScalarFieldContainer[str]
    root_url: str
    path: str
    def __init__(self, runtime_name: _Optional[str] = ..., value_names: _Optional[_Iterable[str]] = ..., root_url: _Optional[str] = ..., path: _Optional[str] = ...) -> None: ...

class BuildResponse(_message.Message):
    __slots__ = ["error", "product"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    PRODUCT_FIELD_NUMBER: _ClassVar[int]
    error: _program_pb2.Error
    product: _build_pb2.BuildArtifact
    def __init__(self, error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ..., product: _Optional[_Union[_build_pb2.BuildArtifact, _Mapping]] = ...) -> None: ...

class RunRequest(_message.Message):
    __slots__ = ["start", "return_load", "return_call"]
    class Start(_message.Message):
        __slots__ = ["run_id", "runtime_name", "root_path", "compiled_data", "values"]
        class ValuesEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: _values_pb2.Value
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
        RUN_ID_FIELD_NUMBER: _ClassVar[int]
        RUNTIME_NAME_FIELD_NUMBER: _ClassVar[int]
        ROOT_PATH_FIELD_NUMBER: _ClassVar[int]
        COMPILED_DATA_FIELD_NUMBER: _ClassVar[int]
        VALUES_FIELD_NUMBER: _ClassVar[int]
        run_id: str
        runtime_name: str
        root_path: str
        compiled_data: bytes
        values: _containers.MessageMap[str, _values_pb2.Value]
        def __init__(self, run_id: _Optional[str] = ..., runtime_name: _Optional[str] = ..., root_path: _Optional[str] = ..., compiled_data: _Optional[bytes] = ..., values: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...
    class LoadReturn(_message.Message):
        __slots__ = ["values", "error"]
        class ValuesEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: _values_pb2.Value
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
        VALUES_FIELD_NUMBER: _ClassVar[int]
        ERROR_FIELD_NUMBER: _ClassVar[int]
        values: _containers.MessageMap[str, _values_pb2.Value]
        error: _program_pb2.Error
        def __init__(self, values: _Optional[_Mapping[str, _values_pb2.Value]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...
    class CallReturn(_message.Message):
        __slots__ = ["values", "error"]
        VALUES_FIELD_NUMBER: _ClassVar[int]
        ERROR_FIELD_NUMBER: _ClassVar[int]
        values: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
        error: _program_pb2.Error
        def __init__(self, values: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...
    START_FIELD_NUMBER: _ClassVar[int]
    RETURN_LOAD_FIELD_NUMBER: _ClassVar[int]
    RETURN_CALL_FIELD_NUMBER: _ClassVar[int]
    start: RunRequest.Start
    return_load: RunRequest.LoadReturn
    return_call: RunRequest.CallReturn
    def __init__(self, start: _Optional[_Union[RunRequest.Start, _Mapping]] = ..., return_load: _Optional[_Union[RunRequest.LoadReturn, _Mapping]] = ..., return_call: _Optional[_Union[RunRequest.CallReturn, _Mapping]] = ...) -> None: ...

class RunResponse(_message.Message):
    __slots__ = ["run_id", "status"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    run_id: str
    status: _run_pb2.RunStatus
    def __init__(self, run_id: _Optional[str] = ..., status: _Optional[_Union[_run_pb2.RunStatus, _Mapping]] = ...) -> None: ...
