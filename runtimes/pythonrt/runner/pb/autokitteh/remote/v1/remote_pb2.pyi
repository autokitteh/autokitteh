from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ContainerConfig(_message.Message):
    __slots__ = ["image"]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    image: str
    def __init__(self, image: _Optional[str] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ["data"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    def __init__(self, data: _Optional[bytes] = ...) -> None: ...

class HealthRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class HealthResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class StartRunnerRequest(_message.Message):
    __slots__ = ["container_config", "build_artifact", "vars", "worker_address"]
    class VarsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CONTAINER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    BUILD_ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    VARS_FIELD_NUMBER: _ClassVar[int]
    WORKER_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    container_config: ContainerConfig
    build_artifact: bytes
    vars: _containers.ScalarMap[str, str]
    worker_address: str
    def __init__(self, container_config: _Optional[_Union[ContainerConfig, _Mapping]] = ..., build_artifact: _Optional[bytes] = ..., vars: _Optional[_Mapping[str, str]] = ..., worker_address: _Optional[str] = ...) -> None: ...

class StartRunnerResponse(_message.Message):
    __slots__ = ["runner_id", "runner_address", "error"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    RUNNER_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    runner_address: str
    error: str
    def __init__(self, runner_id: _Optional[str] = ..., runner_address: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class RunnerHealthRequest(_message.Message):
    __slots__ = ["runner_id"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    def __init__(self, runner_id: _Optional[str] = ...) -> None: ...

class RunnerHealthResponse(_message.Message):
    __slots__ = ["healthy", "error"]
    HEALTHY_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    healthy: bool
    error: str
    def __init__(self, healthy: bool = ..., error: _Optional[str] = ...) -> None: ...

class StopRequest(_message.Message):
    __slots__ = ["runner_id"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    def __init__(self, runner_id: _Optional[str] = ...) -> None: ...

class StopResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class StartRequest(_message.Message):
    __slots__ = ["entry_point", "event"]
    ENTRY_POINT_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    entry_point: str
    event: Event
    def __init__(self, entry_point: _Optional[str] = ..., event: _Optional[_Union[Event, _Mapping]] = ...) -> None: ...

class StartResponse(_message.Message):
    __slots__ = ["error", "traceback"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    TRACEBACK_FIELD_NUMBER: _ClassVar[int]
    error: str
    traceback: _containers.RepeatedCompositeFieldContainer[Frame]
    def __init__(self, error: _Optional[str] = ..., traceback: _Optional[_Iterable[_Union[Frame, _Mapping]]] = ...) -> None: ...

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

class ExecuteRequest(_message.Message):
    __slots__ = ["data"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    def __init__(self, data: _Optional[bytes] = ...) -> None: ...

class ExecuteResponse(_message.Message):
    __slots__ = ["result", "error", "traceback"]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    TRACEBACK_FIELD_NUMBER: _ClassVar[int]
    result: bytes
    error: str
    traceback: _containers.RepeatedCompositeFieldContainer[Frame]
    def __init__(self, result: _Optional[bytes] = ..., error: _Optional[str] = ..., traceback: _Optional[_Iterable[_Union[Frame, _Mapping]]] = ...) -> None: ...

class ActivityReplyRequest(_message.Message):
    __slots__ = ["data", "result", "error"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    result: bytes
    error: str
    def __init__(self, data: _Optional[bytes] = ..., result: _Optional[bytes] = ..., error: _Optional[str] = ...) -> None: ...

class ActivityReplyResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class ExportsRequest(_message.Message):
    __slots__ = ["file_name"]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    file_name: str
    def __init__(self, file_name: _Optional[str] = ...) -> None: ...

class ExportsResponse(_message.Message):
    __slots__ = ["exports", "error"]
    EXPORTS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    exports: _containers.RepeatedScalarFieldContainer[str]
    error: str
    def __init__(self, exports: _Optional[_Iterable[str]] = ..., error: _Optional[str] = ...) -> None: ...

class CallInfo(_message.Message):
    __slots__ = ["function", "args", "kwargs"]
    class KwargsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    FUNCTION_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    KWARGS_FIELD_NUMBER: _ClassVar[int]
    function: str
    args: _containers.RepeatedScalarFieldContainer[str]
    kwargs: _containers.ScalarMap[str, str]
    def __init__(self, function: _Optional[str] = ..., args: _Optional[_Iterable[str]] = ..., kwargs: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ActivityRequest(_message.Message):
    __slots__ = ["runner_id", "data", "call_info"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    CALL_INFO_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    data: bytes
    call_info: CallInfo
    def __init__(self, runner_id: _Optional[str] = ..., data: _Optional[bytes] = ..., call_info: _Optional[_Union[CallInfo, _Mapping]] = ...) -> None: ...

class ActivityResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class DoneRequest(_message.Message):
    __slots__ = ["runner_id", "result", "error", "traceback"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    TRACEBACK_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    result: bytes
    error: str
    traceback: _containers.RepeatedCompositeFieldContainer[Frame]
    def __init__(self, runner_id: _Optional[str] = ..., result: _Optional[bytes] = ..., error: _Optional[str] = ..., traceback: _Optional[_Iterable[_Union[Frame, _Mapping]]] = ...) -> None: ...

class DoneResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class SleepRequest(_message.Message):
    __slots__ = ["runner_id", "duration_ms"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    DURATION_MS_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    duration_ms: int
    def __init__(self, runner_id: _Optional[str] = ..., duration_ms: _Optional[int] = ...) -> None: ...

class SleepResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class SubscribeRequest(_message.Message):
    __slots__ = ["runner_id", "connection", "filter"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    connection: str
    filter: str
    def __init__(self, runner_id: _Optional[str] = ..., connection: _Optional[str] = ..., filter: _Optional[str] = ...) -> None: ...

class SubscribeResponse(_message.Message):
    __slots__ = ["signal_id", "error"]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    signal_id: str
    error: str
    def __init__(self, signal_id: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class NextEventRequest(_message.Message):
    __slots__ = ["runner_id", "signal_ids", "timeout_ms"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    SIGNAL_IDS_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_MS_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    signal_ids: _containers.RepeatedScalarFieldContainer[str]
    timeout_ms: int
    def __init__(self, runner_id: _Optional[str] = ..., signal_ids: _Optional[_Iterable[str]] = ..., timeout_ms: _Optional[int] = ...) -> None: ...

class NextEventResponse(_message.Message):
    __slots__ = ["event", "error"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    event: Event
    error: str
    def __init__(self, event: _Optional[_Union[Event, _Mapping]] = ..., error: _Optional[str] = ...) -> None: ...

class UnsubscribeRequest(_message.Message):
    __slots__ = ["runner_id", "signal_id"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    signal_id: str
    def __init__(self, runner_id: _Optional[str] = ..., signal_id: _Optional[str] = ...) -> None: ...

class UnsubscribeResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class LogRequest(_message.Message):
    __slots__ = ["runner_id", "level", "message"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    LEVEL_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    level: str
    message: str
    def __init__(self, runner_id: _Optional[str] = ..., level: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class LogResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class PrintRequest(_message.Message):
    __slots__ = ["runner_id", "message"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    message: str
    def __init__(self, runner_id: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class PrintResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class IsActiveRunnerRequest(_message.Message):
    __slots__ = ["runner_id"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    def __init__(self, runner_id: _Optional[str] = ...) -> None: ...

class IsActiveRunnerResponse(_message.Message):
    __slots__ = ["is_active", "error"]
    IS_ACTIVE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    is_active: bool
    error: str
    def __init__(self, is_active: bool = ..., error: _Optional[str] = ...) -> None: ...

class RefreshRequest(_message.Message):
    __slots__ = ["runner_id", "integration", "connection"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    integration: str
    connection: str
    def __init__(self, runner_id: _Optional[str] = ..., integration: _Optional[str] = ..., connection: _Optional[str] = ...) -> None: ...

class RefreshResponse(_message.Message):
    __slots__ = ["token", "expires", "error"]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    token: str
    expires: _timestamp_pb2.Timestamp
    error: str
    def __init__(self, token: _Optional[str] = ..., expires: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., error: _Optional[str] = ...) -> None: ...
