from autokitteh.user_code.v1 import user_code_pb2 as _user_code_pb2
from autokitteh.values.v1 import values_pb2 as _values_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CallInfo(_message.Message):
    __slots__ = ["function", "args", "kwargs"]
    class KwargsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    FUNCTION_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    KWARGS_FIELD_NUMBER: _ClassVar[int]
    function: str
    args: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
    kwargs: _containers.MessageMap[str, _values_pb2.Value]
    def __init__(self, function: _Optional[str] = ..., args: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., kwargs: _Optional[_Mapping[str, _values_pb2.Value]] = ...) -> None: ...

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
    result: _values_pb2.Value
    error: str
    traceback: _containers.RepeatedCompositeFieldContainer[_user_code_pb2.Frame]
    def __init__(self, runner_id: _Optional[str] = ..., result: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., error: _Optional[str] = ..., traceback: _Optional[_Iterable[_Union[_user_code_pb2.Frame, _Mapping]]] = ...) -> None: ...

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
    event: _user_code_pb2.Event
    error: str
    def __init__(self, event: _Optional[_Union[_user_code_pb2.Event, _Mapping]] = ..., error: _Optional[str] = ...) -> None: ...

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

class StartSessionRequest(_message.Message):
    __slots__ = ["runner_id", "loc", "data", "memo"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    LOC_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    MEMO_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    loc: str
    data: bytes
    memo: bytes
    def __init__(self, runner_id: _Optional[str] = ..., loc: _Optional[str] = ..., data: _Optional[bytes] = ..., memo: _Optional[bytes] = ...) -> None: ...

class StartSessionResponse(_message.Message):
    __slots__ = ["session_id", "error"]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    error: str
    def __init__(self, session_id: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class EncodeJWTRequest(_message.Message):
    __slots__ = ["runner_id", "payload", "connection", "algorithm"]
    class PayloadEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: int
        def __init__(self, key: _Optional[str] = ..., value: _Optional[int] = ...) -> None: ...
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    ALGORITHM_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    payload: _containers.ScalarMap[str, int]
    connection: str
    algorithm: str
    def __init__(self, runner_id: _Optional[str] = ..., payload: _Optional[_Mapping[str, int]] = ..., connection: _Optional[str] = ..., algorithm: _Optional[str] = ...) -> None: ...

class EncodeJWTResponse(_message.Message):
    __slots__ = ["jwt", "error"]
    JWT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    jwt: str
    error: str
    def __init__(self, jwt: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

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

class HandlerHealthRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class HandlerHealthResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...
