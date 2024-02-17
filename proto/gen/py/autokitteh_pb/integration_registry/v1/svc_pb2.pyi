from autokitteh_pb.integration_registry.v1 import integration_pb2 as _integration_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["integration"]
    INTEGRATION_FIELD_NUMBER: _ClassVar[int]
    integration: _integration_pb2.Integration
    def __init__(self, integration: _Optional[_Union[_integration_pb2.Integration, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["integration_id"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    def __init__(self, integration_id: _Optional[str] = ...) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["integration"]
    INTEGRATION_FIELD_NUMBER: _ClassVar[int]
    integration: _integration_pb2.Integration
    def __init__(self, integration: _Optional[_Union[_integration_pb2.Integration, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["integration_id"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    def __init__(self, integration_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["integration_id"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    def __init__(self, integration_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["integration"]
    INTEGRATION_FIELD_NUMBER: _ClassVar[int]
    integration: _integration_pb2.Integration
    def __init__(self, integration: _Optional[_Union[_integration_pb2.Integration, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["owner_id", "visibility", "api_url"]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    API_URL_FIELD_NUMBER: _ClassVar[int]
    owner_id: str
    visibility: _integration_pb2.Visibility
    api_url: str
    def __init__(self, owner_id: _Optional[str] = ..., visibility: _Optional[_Union[_integration_pb2.Visibility, str]] = ..., api_url: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["integrations"]
    INTEGRATIONS_FIELD_NUMBER: _ClassVar[int]
    integrations: _containers.RepeatedCompositeFieldContainer[_integration_pb2.Integration]
    def __init__(self, integrations: _Optional[_Iterable[_Union[_integration_pb2.Integration, _Mapping]]] = ...) -> None: ...
