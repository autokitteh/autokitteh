from autokitteh_pb.connections.v1 import connection_pb2 as _connection_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["connection"]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    connection: _connection_pb2.Connection
    def __init__(self, connection: _Optional[_Union[_connection_pb2.Connection, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["connection_id"]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["connection"]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    connection: _connection_pb2.Connection
    def __init__(self, connection: _Optional[_Union[_connection_pb2.Connection, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["connection_id"]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["connection_id"]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["connection"]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    connection: _connection_pb2.Connection
    def __init__(self, connection: _Optional[_Union[_connection_pb2.Connection, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["integration_id", "project_id", "integration_token"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_TOKEN_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    project_id: str
    integration_token: str
    def __init__(self, integration_id: _Optional[str] = ..., project_id: _Optional[str] = ..., integration_token: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["connections"]
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    connections: _containers.RepeatedCompositeFieldContainer[_connection_pb2.Connection]
    def __init__(self, connections: _Optional[_Iterable[_Union[_connection_pb2.Connection, _Mapping]]] = ...) -> None: ...
