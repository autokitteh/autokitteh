from autokitteh_pb.orgs.v1 import org_pb2 as _org_pb2
from autokitteh_pb.users.v1 import user_pb2 as _user_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import field_mask_pb2 as _field_mask_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["org"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    org: _org_pb2.Org
    def __init__(self, org: _Optional[_Union[_org_pb2.Org, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["org_id"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    def __init__(self, org_id: _Optional[str] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["org_id", "name"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    name: str
    def __init__(self, org_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["org"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    org: _org_pb2.Org
    def __init__(self, org: _Optional[_Union[_org_pb2.Org, _Mapping]] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["org_id"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    def __init__(self, org_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["org", "field_mask"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    FIELD_MASK_FIELD_NUMBER: _ClassVar[int]
    org: _org_pb2.Org
    field_mask: _field_mask_pb2.FieldMask
    def __init__(self, org: _Optional[_Union[_org_pb2.Org, _Mapping]] = ..., field_mask: _Optional[_Union[_field_mask_pb2.FieldMask, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class AddMemberRequest(_message.Message):
    __slots__ = ["member"]
    MEMBER_FIELD_NUMBER: _ClassVar[int]
    member: _org_pb2.OrgMember
    def __init__(self, member: _Optional[_Union[_org_pb2.OrgMember, _Mapping]] = ...) -> None: ...

class AddMemberResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class RemoveMemberRequest(_message.Message):
    __slots__ = ["user_id", "org_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    org_id: str
    def __init__(self, user_id: _Optional[str] = ..., org_id: _Optional[str] = ...) -> None: ...

class RemoveMemberResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetMemberRequest(_message.Message):
    __slots__ = ["user_id", "org_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    org_id: str
    def __init__(self, user_id: _Optional[str] = ..., org_id: _Optional[str] = ...) -> None: ...

class GetMemberResponse(_message.Message):
    __slots__ = ["member"]
    MEMBER_FIELD_NUMBER: _ClassVar[int]
    member: _org_pb2.OrgMember
    def __init__(self, member: _Optional[_Union[_org_pb2.OrgMember, _Mapping]] = ...) -> None: ...

class ListMembersRequest(_message.Message):
    __slots__ = ["org_id"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    def __init__(self, org_id: _Optional[str] = ...) -> None: ...

class ListMembersResponse(_message.Message):
    __slots__ = ["members", "users"]
    MEMBERS_FIELD_NUMBER: _ClassVar[int]
    USERS_FIELD_NUMBER: _ClassVar[int]
    members: _containers.RepeatedCompositeFieldContainer[_org_pb2.OrgMember]
    users: _containers.RepeatedCompositeFieldContainer[_user_pb2.User]
    def __init__(self, members: _Optional[_Iterable[_Union[_org_pb2.OrgMember, _Mapping]]] = ..., users: _Optional[_Iterable[_Union[_user_pb2.User, _Mapping]]] = ...) -> None: ...

class GetOrgsForUserRequest(_message.Message):
    __slots__ = ["user_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class GetOrgsForUserResponse(_message.Message):
    __slots__ = ["members", "orgs"]
    MEMBERS_FIELD_NUMBER: _ClassVar[int]
    ORGS_FIELD_NUMBER: _ClassVar[int]
    members: _containers.RepeatedCompositeFieldContainer[_org_pb2.OrgMember]
    orgs: _containers.RepeatedCompositeFieldContainer[_org_pb2.Org]
    def __init__(self, members: _Optional[_Iterable[_Union[_org_pb2.OrgMember, _Mapping]]] = ..., orgs: _Optional[_Iterable[_Union[_org_pb2.Org, _Mapping]]] = ...) -> None: ...

class UpdateMemberRequest(_message.Message):
    __slots__ = ["member", "field_mask"]
    MEMBER_FIELD_NUMBER: _ClassVar[int]
    FIELD_MASK_FIELD_NUMBER: _ClassVar[int]
    member: _org_pb2.OrgMember
    field_mask: _field_mask_pb2.FieldMask
    def __init__(self, member: _Optional[_Union[_org_pb2.OrgMember, _Mapping]] = ..., field_mask: _Optional[_Union[_field_mask_pb2.FieldMask, _Mapping]] = ...) -> None: ...

class UpdateMemberResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
