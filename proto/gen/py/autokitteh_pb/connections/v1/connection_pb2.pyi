from autokitteh_pb.common.v1 import status_pb2 as _status_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Connection(_message.Message):
    __slots__ = ["connection_id", "integration_id", "project_id", "name", "status", "capabilities", "links"]
    class LinksEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CAPABILITIES_FIELD_NUMBER: _ClassVar[int]
    LINKS_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    integration_id: str
    project_id: str
    name: str
    status: _status_pb2.Status
    capabilities: Capabilities
    links: _containers.ScalarMap[str, str]
    def __init__(self, connection_id: _Optional[str] = ..., integration_id: _Optional[str] = ..., project_id: _Optional[str] = ..., name: _Optional[str] = ..., status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ..., capabilities: _Optional[_Union[Capabilities, _Mapping]] = ..., links: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Capabilities(_message.Message):
    __slots__ = ["supports_connection_test", "supports_connection_init", "requires_connection_init"]
    SUPPORTS_CONNECTION_TEST_FIELD_NUMBER: _ClassVar[int]
    SUPPORTS_CONNECTION_INIT_FIELD_NUMBER: _ClassVar[int]
    REQUIRES_CONNECTION_INIT_FIELD_NUMBER: _ClassVar[int]
    supports_connection_test: bool
    supports_connection_init: bool
    requires_connection_init: bool
    def __init__(self, supports_connection_test: bool = ..., supports_connection_init: bool = ..., requires_connection_init: bool = ...) -> None: ...
