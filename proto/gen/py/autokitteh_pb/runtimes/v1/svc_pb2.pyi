from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.runtimes.v1 import build_pb2 as _build_pb2
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
    __slots__ = ["resources", "symbols", "memo"]
    class ResourcesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    class MemoEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    MEMO_FIELD_NUMBER: _ClassVar[int]
    resources: _containers.ScalarMap[str, bytes]
    symbols: _containers.RepeatedScalarFieldContainer[str]
    memo: _containers.ScalarMap[str, str]
    def __init__(self, resources: _Optional[_Mapping[str, bytes]] = ..., symbols: _Optional[_Iterable[str]] = ..., memo: _Optional[_Mapping[str, str]] = ...) -> None: ...

class BuildResponse(_message.Message):
    __slots__ = ["build_file", "error"]
    BUILD_FILE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    build_file: bytes
    error: _program_pb2.Error
    def __init__(self, build_file: _Optional[bytes] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...

class Build1Request(_message.Message):
    __slots__ = ["resources", "symbols", "path", "runtime_name"]
    class ResourcesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    SYMBOLS_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_NAME_FIELD_NUMBER: _ClassVar[int]
    resources: _containers.ScalarMap[str, bytes]
    symbols: _containers.RepeatedScalarFieldContainer[str]
    path: str
    runtime_name: str
    def __init__(self, resources: _Optional[_Mapping[str, bytes]] = ..., symbols: _Optional[_Iterable[str]] = ..., path: _Optional[str] = ..., runtime_name: _Optional[str] = ...) -> None: ...

class Build1Response(_message.Message):
    __slots__ = ["artifact", "error"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    artifact: _build_pb2.Artifact
    error: _program_pb2.Error
    def __init__(self, artifact: _Optional[_Union[_build_pb2.Artifact, _Mapping]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...

class RunRequest(_message.Message):
    __slots__ = ["run_id", "artifact", "path", "globals"]
    class GlobalsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    GLOBALS_FIELD_NUMBER: _ClassVar[int]
    run_id: str
    artifact: bytes
    path: str
    globals: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, run_id: _Optional[str] = ..., artifact: _Optional[bytes] = ..., path: _Optional[str] = ..., globals: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class RunResponse(_message.Message):
    __slots__ = ["print", "error", "result"]
    class ResultEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    PRINT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    print: str
    error: _program_pb2.Error
    result: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, print: _Optional[str] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ..., result: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class BidiRunCall(_message.Message):
    __slots__ = ["value", "args", "kwargs"]
    class KwargsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    VALUE_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    KWARGS_FIELD_NUMBER: _ClassVar[int]
    value: _values_pb2.Value
    args: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
    kwargs: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., args: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., kwargs: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class BidiRunCallReturn(_message.Message):
    __slots__ = ["value", "error"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    value: _values_pb2.Value
    error: _program_pb2.Error
    def __init__(self, value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...

class BidiRunLoadReturn(_message.Message):
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

class BidiRunRequest(_message.Message):
    __slots__ = ["start", "start1", "call", "call_return", "load_return", "new_run_id_value"]
    class StartData(_message.Message):
        __slots__ = ["run_id", "globals", "path"]
        class GlobalsEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: _values_pb2.Value
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
        RUN_ID_FIELD_NUMBER: _ClassVar[int]
        GLOBALS_FIELD_NUMBER: _ClassVar[int]
        PATH_FIELD_NUMBER: _ClassVar[int]
        run_id: str
        globals: _containers.MessageMap[str, _values_pb2.Value]
        path: str
        def __init__(self, run_id: _Optional[str] = ..., globals: _Optional[_Mapping[str, _values_pb2.Value]] = ..., path: _Optional[str] = ...) -> None: ...
    class Start(_message.Message):
        __slots__ = ["build_file", "data"]
        BUILD_FILE_FIELD_NUMBER: _ClassVar[int]
        DATA_FIELD_NUMBER: _ClassVar[int]
        build_file: bytes
        data: BidiRunRequest.StartData
        def __init__(self, build_file: _Optional[bytes] = ..., data: _Optional[_Union[BidiRunRequest.StartData, _Mapping]] = ...) -> None: ...
    class Start1(_message.Message):
        __slots__ = ["runtime_name", "artifact", "data"]
        RUNTIME_NAME_FIELD_NUMBER: _ClassVar[int]
        ARTIFACT_FIELD_NUMBER: _ClassVar[int]
        DATA_FIELD_NUMBER: _ClassVar[int]
        runtime_name: str
        artifact: _build_pb2.Artifact
        data: BidiRunRequest.StartData
        def __init__(self, runtime_name: _Optional[str] = ..., artifact: _Optional[_Union[_build_pb2.Artifact, _Mapping]] = ..., data: _Optional[_Union[BidiRunRequest.StartData, _Mapping]] = ...) -> None: ...
    class Call(_message.Message):
        __slots__ = ["value", "args", "kwargs"]
        class KwargsEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: _values_pb2.Value
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
        VALUE_FIELD_NUMBER: _ClassVar[int]
        ARGS_FIELD_NUMBER: _ClassVar[int]
        KWARGS_FIELD_NUMBER: _ClassVar[int]
        value: _values_pb2.Value
        args: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
        kwargs: _containers.MessageMap[str, _values_pb2.Value]
        def __init__(self, value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., args: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., kwargs: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...
    class NewRunIDValue(_message.Message):
        __slots__ = ["run_id"]
        RUN_ID_FIELD_NUMBER: _ClassVar[int]
        run_id: str
        def __init__(self, run_id: _Optional[str] = ...) -> None: ...
    START_FIELD_NUMBER: _ClassVar[int]
    START1_FIELD_NUMBER: _ClassVar[int]
    CALL_FIELD_NUMBER: _ClassVar[int]
    CALL_RETURN_FIELD_NUMBER: _ClassVar[int]
    LOAD_RETURN_FIELD_NUMBER: _ClassVar[int]
    NEW_RUN_ID_VALUE_FIELD_NUMBER: _ClassVar[int]
    start: BidiRunRequest.Start
    start1: BidiRunRequest.Start1
    call: BidiRunCall
    call_return: BidiRunCallReturn
    load_return: BidiRunLoadReturn
    new_run_id_value: BidiRunRequest.NewRunIDValue
    def __init__(self, start: _Optional[_Union[BidiRunRequest.Start, _Mapping]] = ..., start1: _Optional[_Union[BidiRunRequest.Start1, _Mapping]] = ..., call: _Optional[_Union[BidiRunCall, _Mapping]] = ..., call_return: _Optional[_Union[BidiRunCallReturn, _Mapping]] = ..., load_return: _Optional[_Union[BidiRunLoadReturn, _Mapping]] = ..., new_run_id_value: _Optional[_Union[BidiRunRequest.NewRunIDValue, _Mapping]] = ...) -> None: ...

class BidiRunResponse(_message.Message):
    __slots__ = ["print", "call", "call_return", "load", "start_return", "new_run_id"]
    class Print(_message.Message):
        __slots__ = ["text"]
        TEXT_FIELD_NUMBER: _ClassVar[int]
        text: str
        def __init__(self, text: _Optional[str] = ...) -> None: ...
    class Load(_message.Message):
        __slots__ = ["path"]
        PATH_FIELD_NUMBER: _ClassVar[int]
        path: str
        def __init__(self, path: _Optional[str] = ...) -> None: ...
    class NewRunID(_message.Message):
        __slots__ = []
        def __init__(self) -> None: ...
    PRINT_FIELD_NUMBER: _ClassVar[int]
    CALL_FIELD_NUMBER: _ClassVar[int]
    CALL_RETURN_FIELD_NUMBER: _ClassVar[int]
    LOAD_FIELD_NUMBER: _ClassVar[int]
    START_RETURN_FIELD_NUMBER: _ClassVar[int]
    NEW_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    print: BidiRunResponse.Print
    call: BidiRunCall
    call_return: BidiRunCallReturn
    load: BidiRunResponse.Load
    start_return: BidiRunLoadReturn
    new_run_id: BidiRunResponse.NewRunID
    def __init__(self, print: _Optional[_Union[BidiRunResponse.Print, _Mapping]] = ..., call: _Optional[_Union[BidiRunCall, _Mapping]] = ..., call_return: _Optional[_Union[BidiRunCallReturn, _Mapping]] = ..., load: _Optional[_Union[BidiRunResponse.Load, _Mapping]] = ..., start_return: _Optional[_Union[BidiRunLoadReturn, _Mapping]] = ..., new_run_id: _Optional[_Union[BidiRunResponse.NewRunID, _Mapping]] = ...) -> None: ...
