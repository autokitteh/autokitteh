from pb.autokitteh.user_code.v1 import user_code_pb2 as _user_code_pb2
from pb.autokitteh.values.v1 import values_pb2 as _values_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

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

class StartRequest(_message.Message):
    __slots__ = ["entry_point", "event"]
    ENTRY_POINT_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    entry_point: str
    event: _user_code_pb2.Event
    def __init__(self, entry_point: _Optional[str] = ..., event: _Optional[_Union[_user_code_pb2.Event, _Mapping]] = ...) -> None: ...

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
    result: _values_pb2.Value
    error: str
    traceback: _containers.RepeatedCompositeFieldContainer[_user_code_pb2.Frame]
    def __init__(self, result: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., error: _Optional[str] = ..., traceback: _Optional[_Iterable[_Union[_user_code_pb2.Frame, _Mapping]]] = ...) -> None: ...

class StartResponse(_message.Message):
    __slots__ = ["error", "traceback"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    TRACEBACK_FIELD_NUMBER: _ClassVar[int]
    error: str
    traceback: _containers.RepeatedCompositeFieldContainer[_user_code_pb2.Frame]
    def __init__(self, error: _Optional[str] = ..., traceback: _Optional[_Iterable[_Union[_user_code_pb2.Frame, _Mapping]]] = ...) -> None: ...

class ActivityReplyRequest(_message.Message):
    __slots__ = ["result", "error"]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    result: _values_pb2.Value
    error: str
    def __init__(self, result: _Optional[_Union[_values_pb2.Value, _Mapping]] = ..., error: _Optional[str] = ...) -> None: ...

class ActivityReplyResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class RunnerHealthRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class RunnerHealthResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...
