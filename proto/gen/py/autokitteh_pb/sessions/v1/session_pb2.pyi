from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.values.v1 import values_pb2 as _values_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SessionStateType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SESSION_STATE_TYPE_UNSPECIFIED: _ClassVar[SessionStateType]
    SESSION_STATE_TYPE_CREATED: _ClassVar[SessionStateType]
    SESSION_STATE_TYPE_RUNNING: _ClassVar[SessionStateType]
    SESSION_STATE_TYPE_ERROR: _ClassVar[SessionStateType]
    SESSION_STATE_TYPE_COMPLETED: _ClassVar[SessionStateType]
SESSION_STATE_TYPE_UNSPECIFIED: SessionStateType
SESSION_STATE_TYPE_CREATED: SessionStateType
SESSION_STATE_TYPE_RUNNING: SessionStateType
SESSION_STATE_TYPE_ERROR: SessionStateType
SESSION_STATE_TYPE_COMPLETED: SessionStateType

class SessionState(_message.Message):
    __slots__ = ["t", "created", "running", "error", "completed"]
    class Created(_message.Message):
        __slots__ = []
        def __init__(self) -> None: ...
    class Running(_message.Message):
        __slots__ = ["run_id", "call"]
        RUN_ID_FIELD_NUMBER: _ClassVar[int]
        CALL_FIELD_NUMBER: _ClassVar[int]
        run_id: str
        call: _values_pb2.Value
        def __init__(self, run_id: _Optional[str] = ..., call: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    class Error(_message.Message):
        __slots__ = ["prints", "error"]
        PRINTS_FIELD_NUMBER: _ClassVar[int]
        ERROR_FIELD_NUMBER: _ClassVar[int]
        prints: _containers.RepeatedScalarFieldContainer[str]
        error: _program_pb2.Error
        def __init__(self, prints: _Optional[_Iterable[str]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...
    class Completed(_message.Message):
        __slots__ = ["prints", "exports", "return_value"]
        class ExportsEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: _values_pb2.Value
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
        PRINTS_FIELD_NUMBER: _ClassVar[int]
        EXPORTS_FIELD_NUMBER: _ClassVar[int]
        RETURN_VALUE_FIELD_NUMBER: _ClassVar[int]
        prints: _containers.RepeatedScalarFieldContainer[str]
        exports: _containers.MessageMap[str, _values_pb2.Value]
        return_value: _values_pb2.Value
        def __init__(self, prints: _Optional[_Iterable[str]] = ..., exports: _Optional[_Mapping[str, _values_pb2.Value]] = ..., return_value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    T_FIELD_NUMBER: _ClassVar[int]
    CREATED_FIELD_NUMBER: _ClassVar[int]
    RUNNING_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    COMPLETED_FIELD_NUMBER: _ClassVar[int]
    t: _timestamp_pb2.Timestamp
    created: SessionState.Created
    running: SessionState.Running
    error: SessionState.Error
    completed: SessionState.Completed
    def __init__(self, t: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., created: _Optional[_Union[SessionState.Created, _Mapping]] = ..., running: _Optional[_Union[SessionState.Running, _Mapping]] = ..., error: _Optional[_Union[SessionState.Error, _Mapping]] = ..., completed: _Optional[_Union[SessionState.Completed, _Mapping]] = ...) -> None: ...

class Call(_message.Message):
    __slots__ = ["spec", "attempts"]
    class Spec(_message.Message):
        __slots__ = ["function", "args", "kwargs", "seq"]
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
        SEQ_FIELD_NUMBER: _ClassVar[int]
        function: _values_pb2.Value
        args: _containers.RepeatedCompositeFieldContainer[_values_pb2.Value]
        kwargs: _containers.MessageMap[str, _values_pb2.Value]
        seq: int
        def __init__(self, function: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., args: _Optional[_Iterable[_Union[_values_pb2.Value, _Mapping]]] = ..., kwargs: _Optional[_Mapping[str, _values_pb2.Value]] = ..., seq: _Optional[int] = ...) -> None: ...
    class Attempt(_message.Message):
        __slots__ = ["start", "complete"]
        class Result(_message.Message):
            __slots__ = ["value", "error"]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            ERROR_FIELD_NUMBER: _ClassVar[int]
            value: _values_pb2.Value
            error: _program_pb2.Error
            def __init__(self, value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...
        class Start(_message.Message):
            __slots__ = ["started_at", "num"]
            STARTED_AT_FIELD_NUMBER: _ClassVar[int]
            NUM_FIELD_NUMBER: _ClassVar[int]
            started_at: _timestamp_pb2.Timestamp
            num: int
            def __init__(self, started_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., num: _Optional[int] = ...) -> None: ...
        class Complete(_message.Message):
            __slots__ = ["completed_at", "retry_interval", "is_last", "result"]
            COMPLETED_AT_FIELD_NUMBER: _ClassVar[int]
            RETRY_INTERVAL_FIELD_NUMBER: _ClassVar[int]
            IS_LAST_FIELD_NUMBER: _ClassVar[int]
            RESULT_FIELD_NUMBER: _ClassVar[int]
            completed_at: _timestamp_pb2.Timestamp
            retry_interval: _duration_pb2.Duration
            is_last: bool
            result: Call.Attempt.Result
            def __init__(self, completed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., retry_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., is_last: bool = ..., result: _Optional[_Union[Call.Attempt.Result, _Mapping]] = ...) -> None: ...
        START_FIELD_NUMBER: _ClassVar[int]
        COMPLETE_FIELD_NUMBER: _ClassVar[int]
        start: Call.Attempt.Start
        complete: Call.Attempt.Complete
        def __init__(self, start: _Optional[_Union[Call.Attempt.Start, _Mapping]] = ..., complete: _Optional[_Union[Call.Attempt.Complete, _Mapping]] = ...) -> None: ...
    SPEC_FIELD_NUMBER: _ClassVar[int]
    ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    spec: Call.Spec
    attempts: _containers.RepeatedCompositeFieldContainer[Call.Attempt]
    def __init__(self, spec: _Optional[_Union[Call.Spec, _Mapping]] = ..., attempts: _Optional[_Iterable[_Union[Call.Attempt, _Mapping]]] = ...) -> None: ...

class SessionLogRecord(_message.Message):
    __slots__ = ["t", "print", "call_spec", "call_attempt_start", "call_attempt_complete", "state"]
    T_FIELD_NUMBER: _ClassVar[int]
    PRINT_FIELD_NUMBER: _ClassVar[int]
    CALL_SPEC_FIELD_NUMBER: _ClassVar[int]
    CALL_ATTEMPT_START_FIELD_NUMBER: _ClassVar[int]
    CALL_ATTEMPT_COMPLETE_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    t: _timestamp_pb2.Timestamp
    print: str
    call_spec: Call.Spec
    call_attempt_start: Call.Attempt.Start
    call_attempt_complete: Call.Attempt.Complete
    state: SessionState
    def __init__(self, t: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., print: _Optional[str] = ..., call_spec: _Optional[_Union[Call.Spec, _Mapping]] = ..., call_attempt_start: _Optional[_Union[Call.Attempt.Start, _Mapping]] = ..., call_attempt_complete: _Optional[_Union[Call.Attempt.Complete, _Mapping]] = ..., state: _Optional[_Union[SessionState, _Mapping]] = ...) -> None: ...

class SessionLog(_message.Message):
    __slots__ = ["records"]
    RECORDS_FIELD_NUMBER: _ClassVar[int]
    records: _containers.RepeatedCompositeFieldContainer[SessionLogRecord]
    def __init__(self, records: _Optional[_Iterable[_Union[SessionLogRecord, _Mapping]]] = ...) -> None: ...

class Session(_message.Message):
    __slots__ = ["session_id", "deployment_id", "event_id", "entrypoint", "inputs", "parent_session_id", "memo", "created_at", "updated_at", "state"]
    class InputsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _values_pb2.Value
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_values_pb2.Value, _Mapping]] = ...) -> None: ...
    class MemoEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ENTRYPOINT_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    PARENT_SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    MEMO_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    deployment_id: str
    event_id: str
    entrypoint: _program_pb2.CodeLocation
    inputs: _containers.MessageMap[str, _values_pb2.Value]
    parent_session_id: str
    memo: _containers.ScalarMap[str, str]
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    state: SessionStateType
    def __init__(self, session_id: _Optional[str] = ..., deployment_id: _Optional[str] = ..., event_id: _Optional[str] = ..., entrypoint: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., inputs: _Optional[_Mapping[str, _values_pb2.Value]] = ..., parent_session_id: _Optional[str] = ..., memo: _Optional[_Mapping[str, str]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., state: _Optional[_Union[SessionStateType, str]] = ...) -> None: ...
