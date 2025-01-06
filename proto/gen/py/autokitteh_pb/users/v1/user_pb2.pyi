from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class UserStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    USER_STATUS_UNSPECIFIED: _ClassVar[UserStatus]
    USER_STATUS_ACTIVE: _ClassVar[UserStatus]
    USER_STATUS_INVITED: _ClassVar[UserStatus]
    USER_STATUS_DISABLED: _ClassVar[UserStatus]
USER_STATUS_UNSPECIFIED: UserStatus
USER_STATUS_ACTIVE: UserStatus
USER_STATUS_INVITED: UserStatus
USER_STATUS_DISABLED: UserStatus

class User(_message.Message):
    __slots__ = ["user_id", "email", "display_name", "disabled", "default_org_id", "status"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_NAME_FIELD_NUMBER: _ClassVar[int]
    DISABLED_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_ORG_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    email: str
    display_name: str
    disabled: bool
    default_org_id: str
    status: UserStatus
    def __init__(self, user_id: _Optional[str] = ..., email: _Optional[str] = ..., display_name: _Optional[str] = ..., disabled: bool = ..., default_org_id: _Optional[str] = ..., status: _Optional[_Union[UserStatus, str]] = ...) -> None: ...
