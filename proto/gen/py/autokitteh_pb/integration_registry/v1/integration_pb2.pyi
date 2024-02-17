from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Visibility(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    VISIBILITY_UNSPECIFIED: _ClassVar[Visibility]
    VISIBILITY_PRIVATE: _ClassVar[Visibility]
    VISIBILITY_INTERNAL: _ClassVar[Visibility]
    VISIBILITY_PUBLIC: _ClassVar[Visibility]
VISIBILITY_UNSPECIFIED: Visibility
VISIBILITY_PRIVATE: Visibility
VISIBILITY_INTERNAL: Visibility
VISIBILITY_PUBLIC: Visibility

class Integration(_message.Message):
    __slots__ = ["integration_id", "name", "owner_id", "visibility", "api_url", "display_name", "description", "logo_url", "homepage_url", "connect_url", "api_key", "signing_key"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    API_URL_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    LOGO_URL_FIELD_NUMBER: _ClassVar[int]
    HOMEPAGE_URL_FIELD_NUMBER: _ClassVar[int]
    CONNECT_URL_FIELD_NUMBER: _ClassVar[int]
    API_KEY_FIELD_NUMBER: _ClassVar[int]
    SIGNING_KEY_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    name: str
    owner_id: str
    visibility: Visibility
    api_url: str
    display_name: str
    description: str
    logo_url: str
    homepage_url: str
    connect_url: str
    api_key: str
    signing_key: str
    def __init__(self, integration_id: _Optional[str] = ..., name: _Optional[str] = ..., owner_id: _Optional[str] = ..., visibility: _Optional[_Union[Visibility, str]] = ..., api_url: _Optional[str] = ..., display_name: _Optional[str] = ..., description: _Optional[str] = ..., logo_url: _Optional[str] = ..., homepage_url: _Optional[str] = ..., connect_url: _Optional[str] = ..., api_key: _Optional[str] = ..., signing_key: _Optional[str] = ...) -> None: ...
