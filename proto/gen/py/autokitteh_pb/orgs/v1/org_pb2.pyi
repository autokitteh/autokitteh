from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class OrgMemberStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ORG_MEMBER_STATUS_UNSPECIFIED: _ClassVar[OrgMemberStatus]
    ORG_MEMBER_STATUS_ACTIVE: _ClassVar[OrgMemberStatus]
    ORG_MEMBER_STATUS_INVITED: _ClassVar[OrgMemberStatus]
    ORG_MEMBER_STATUS_DECLINED: _ClassVar[OrgMemberStatus]
ORG_MEMBER_STATUS_UNSPECIFIED: OrgMemberStatus
ORG_MEMBER_STATUS_ACTIVE: OrgMemberStatus
ORG_MEMBER_STATUS_INVITED: OrgMemberStatus
ORG_MEMBER_STATUS_DECLINED: OrgMemberStatus

class Org(_message.Message):
    __slots__ = ["org_id", "display_name"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_NAME_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    display_name: str
    def __init__(self, org_id: _Optional[str] = ..., display_name: _Optional[str] = ...) -> None: ...

class OrgMember(_message.Message):
    __slots__ = ["user_id", "org_id", "status"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    org_id: str
    status: OrgMemberStatus
    def __init__(self, user_id: _Optional[str] = ..., org_id: _Optional[str] = ..., status: _Optional[_Union[OrgMemberStatus, str]] = ...) -> None: ...
