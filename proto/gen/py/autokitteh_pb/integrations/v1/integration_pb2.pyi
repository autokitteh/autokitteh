from autokitteh_pb.program.v1 import module_pb2 as _module_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Integration(_message.Message):
    __slots__ = ["integration_id", "unique_name", "display_name", "description", "logo_url", "user_links", "connection_url", "module"]
    class UserLinksEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    UNIQUE_NAME_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    LOGO_URL_FIELD_NUMBER: _ClassVar[int]
    USER_LINKS_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_URL_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    unique_name: str
    display_name: str
    description: str
    logo_url: str
    user_links: _containers.ScalarMap[str, str]
    connection_url: str
    module: _module_pb2.Module
    def __init__(self, integration_id: _Optional[str] = ..., unique_name: _Optional[str] = ..., display_name: _Optional[str] = ..., description: _Optional[str] = ..., logo_url: _Optional[str] = ..., user_links: _Optional[_Mapping[str, str]] = ..., connection_url: _Optional[str] = ..., module: _Optional[_Union[_module_pb2.Module, _Mapping]] = ...) -> None: ...
