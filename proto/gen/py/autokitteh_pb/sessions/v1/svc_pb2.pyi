from autokitteh_pb.sessions.v1 import session_pb2 as _session_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class StartRequest(_message.Message):
    __slots__ = ["session", "json_inputs"]
    class JsonInputsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    SESSION_FIELD_NUMBER: _ClassVar[int]
    JSON_INPUTS_FIELD_NUMBER: _ClassVar[int]
    session: _session_pb2.Session
    json_inputs: _containers.ScalarMap[str, str]
    def __init__(self, session: _Optional[_Union[_session_pb2.Session, _Mapping]] = ..., json_inputs: _Optional[_Mapping[str, str]] = ...) -> None: ...

class StartResponse(_message.Message):
    __slots__ = ["session_id"]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class StopRequest(_message.Message):
    __slots__ = ["session_id", "reason", "terminate"]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    TERMINATE_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    reason: str
    terminate: bool
    def __init__(self, session_id: _Optional[str] = ..., reason: _Optional[str] = ..., terminate: bool = ...) -> None: ...

class StopResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["deployment_id", "project_id", "event_id", "build_id", "state_type", "count_only", "page_size", "skip", "page_token"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_TYPE_FIELD_NUMBER: _ClassVar[int]
    COUNT_ONLY_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    SKIP_FIELD_NUMBER: _ClassVar[int]
    PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    project_id: str
    event_id: str
    build_id: str
    state_type: _session_pb2.SessionStateType
    count_only: bool
    page_size: int
    skip: int
    page_token: str
    def __init__(self, deployment_id: _Optional[str] = ..., project_id: _Optional[str] = ..., event_id: _Optional[str] = ..., build_id: _Optional[str] = ..., state_type: _Optional[_Union[_session_pb2.SessionStateType, str]] = ..., count_only: bool = ..., page_size: _Optional[int] = ..., skip: _Optional[int] = ..., page_token: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["sessions", "count", "next_page_token"]
    SESSIONS_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    NEXT_PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    sessions: _containers.RepeatedCompositeFieldContainer[_session_pb2.Session]
    count: int
    next_page_token: str
    def __init__(self, sessions: _Optional[_Iterable[_Union[_session_pb2.Session, _Mapping]]] = ..., count: _Optional[int] = ..., next_page_token: _Optional[str] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["session_id", "json_values"]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    JSON_VALUES_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    json_values: bool
    def __init__(self, session_id: _Optional[str] = ..., json_values: bool = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["session"]
    SESSION_FIELD_NUMBER: _ClassVar[int]
    session: _session_pb2.Session
    def __init__(self, session: _Optional[_Union[_session_pb2.Session, _Mapping]] = ...) -> None: ...

class GetLogRequest(_message.Message):
    __slots__ = ["session_id", "json_values", "types", "ascending", "page_size", "skip", "page_token"]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    JSON_VALUES_FIELD_NUMBER: _ClassVar[int]
    TYPES_FIELD_NUMBER: _ClassVar[int]
    ASCENDING_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    SKIP_FIELD_NUMBER: _ClassVar[int]
    PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    json_values: bool
    types: _session_pb2.SessionLogRecord.Type
    ascending: bool
    page_size: int
    skip: int
    page_token: str
    def __init__(self, session_id: _Optional[str] = ..., json_values: bool = ..., types: _Optional[_Union[_session_pb2.SessionLogRecord.Type, str]] = ..., ascending: bool = ..., page_size: _Optional[int] = ..., skip: _Optional[int] = ..., page_token: _Optional[str] = ...) -> None: ...

class GetLogResponse(_message.Message):
    __slots__ = ["log", "count", "next_page_token"]
    LOG_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    NEXT_PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    log: _session_pb2.SessionLog
    count: int
    next_page_token: str
    def __init__(self, log: _Optional[_Union[_session_pb2.SessionLog, _Mapping]] = ..., count: _Optional[int] = ..., next_page_token: _Optional[str] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["session_id"]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
