from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CallWait(_message.Message):
    __slots__ = ["call", "args", "kwargs"]
    class KwargsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    CALL_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    KWARGS_FIELD_NUMBER: _ClassVar[int]
    call: _values_pb2.Value
    args: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
    kwargs: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, call: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., args: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., kwargs: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

class RunStatus(_message.Message):
    __slots__ = ["idle", "running", "load_wait", "call_wait", "completed", "error"]
    class Idle(_message.Message):
        __slots__ = []
        def __init__(self) -> None: ...
    class Running(_message.Message):
        __slots__ = []
        def __init__(self) -> None: ...
    class LoadWait(_message.Message):
        __slots__ = ["path"]
        PATH_FIELD_NUMBER: _ClassVar[int]
        path: str
        def __init__(self, path: _Optional[str] = ...) -> None: ...
    class Completed(_message.Message):
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
    class Error(_message.Message):
        __slots__ = ["errors"]
        ERRORS_FIELD_NUMBER: _ClassVar[int]
        errors: _containers.RepeatedCompositeFieldContainer[_program_pb2.Error]
        def __init__(self, errors: _Optional[_Iterable[_Union[_program_pb2.Error, _Mapping]]] = ...) -> None: ...
    IDLE_FIELD_NUMBER: _ClassVar[int]
    RUNNING_FIELD_NUMBER: _ClassVar[int]
    LOAD_WAIT_FIELD_NUMBER: _ClassVar[int]
    CALL_WAIT_FIELD_NUMBER: _ClassVar[int]
    COMPLETED_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    idle: RunStatus.Idle
    running: RunStatus.Running
    load_wait: RunStatus.LoadWait
    call_wait: CallWait
    completed: RunStatus.Completed
    error: RunStatus.Error
    def __init__(self, idle: _Optional[_Union[RunStatus.Idle, _Mapping]] = ..., running: _Optional[_Union[RunStatus.Running, _Mapping]] = ..., load_wait: _Optional[_Union[RunStatus.LoadWait, _Mapping]] = ..., call_wait: _Optional[_Union[CallWait, _Mapping]] = ..., completed: _Optional[_Union[RunStatus.Completed, _Mapping]] = ..., error: _Optional[_Union[RunStatus.Error, _Mapping]] = ...) -> None: ...
