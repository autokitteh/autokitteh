from autokitteh_pb.mappings.v1 import mapping_pb2 as _mapping_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["mapping"]
    MAPPING_FIELD_NUMBER: _ClassVar[int]
    mapping: _mapping_pb2.Mapping
    def __init__(self, mapping: _Optional[_Union[_mapping_pb2.Mapping, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["mapping_id"]
    MAPPING_ID_FIELD_NUMBER: _ClassVar[int]
    mapping_id: str
    def __init__(self, mapping_id: _Optional[str] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["mapping_id"]
    MAPPING_ID_FIELD_NUMBER: _ClassVar[int]
    mapping_id: str
    def __init__(self, mapping_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["mapping_id"]
    MAPPING_ID_FIELD_NUMBER: _ClassVar[int]
    mapping_id: str
    def __init__(self, mapping_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["mapping"]
    MAPPING_FIELD_NUMBER: _ClassVar[int]
    mapping: _mapping_pb2.Mapping
    def __init__(self, mapping: _Optional[_Union[_mapping_pb2.Mapping, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["env_id", "connection_id"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    connection_id: str
    def __init__(self, env_id: _Optional[str] = ..., connection_id: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["mappings"]
    MAPPINGS_FIELD_NUMBER: _ClassVar[int]
    mappings: _containers.RepeatedCompositeFieldContainer[_mapping_pb2.Mapping]
    def __init__(self, mappings: _Optional[_Iterable[_Union[_mapping_pb2.Mapping, _Mapping]]] = ...) -> None: ...
