from autokitteh_pb.common.v1 import status_pb2 as _status_pb2
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
    __slots__ = ["integration_id", "project_id", "status_code", "org_id"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_CODE_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    project_id: str
    status_code: _status_pb2.Status.Code
    org_id: str
    def __init__(self, integration_id: _Optional[str] = ..., project_id: _Optional[str] = ..., status_code: _Optional[_Union[_status_pb2.Status.Code, str]] = ..., org_id: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["connections"]
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    connections: _containers.RepeatedCompositeFieldContainer[_connection_pb2.Connection]
    def __init__(self, connections: _Optional[_Iterable[_Union[_connection_pb2.Connection, _Mapping]]] = ...) -> None: ...

class TestRequest(_message.Message):
    __slots__ = ["connection_id"]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class TestResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: _status_pb2.Status
    def __init__(self, status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...

class RefreshStatusRequest(_message.Message):
    __slots__ = ["connection_id"]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class RefreshStatusResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: _status_pb2.Status
    def __init__(self, status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...
