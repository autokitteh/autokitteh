from autokitteh_pb.builds.v1 import build_pb2 as _build_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetRequest(_message.Message):
    __slots__ = ["build_id"]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    def __init__(self, build_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["build"]
    BUILD_FIELD_NUMBER: _ClassVar[int]
    build: _build_pb2.Build
    def __init__(self, build: _Optional[_Union[_build_pb2.Build, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["limit"]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    limit: int
    def __init__(self, limit: _Optional[int] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["builds"]
    BUILDS_FIELD_NUMBER: _ClassVar[int]
    builds: _containers.RepeatedCompositeFieldContainer[_build_pb2.Build]
    def __init__(self, builds: _Optional[_Iterable[_Union[_build_pb2.Build, _Mapping]]] = ...) -> None: ...

class SaveRequest(_message.Message):
    __slots__ = ["build", "data"]
    BUILD_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    build: _build_pb2.Build
    data: bytes
    def __init__(self, build: _Optional[_Union[_build_pb2.Build, _Mapping]] = ..., data: _Optional[bytes] = ...) -> None: ...

class SaveResponse(_message.Message):
    __slots__ = ["build_id"]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    def __init__(self, build_id: _Optional[str] = ...) -> None: ...

class RemoveRequest(_message.Message):
    __slots__ = ["build_id"]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    def __init__(self, build_id: _Optional[str] = ...) -> None: ...

class RemoveResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DownloadRequest(_message.Message):
    __slots__ = ["build_id"]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    def __init__(self, build_id: _Optional[str] = ...) -> None: ...

class DownloadResponse(_message.Message):
    __slots__ = ["data"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    def __init__(self, data: _Optional[bytes] = ...) -> None: ...
