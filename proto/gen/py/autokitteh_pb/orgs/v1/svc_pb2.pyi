from autokitteh_pb.orgs.v1 import org_pb2 as _org_pb2
from buf.validate import validate_pb2 as _validate_pb2
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
    __slots__ = ["org_id"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    def __init__(self, org_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["org"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    org: _org_pb2.Org
    def __init__(self, org: _Optional[_Union[_org_pb2.Org, _Mapping]] = ...) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["org"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    org: _org_pb2.Org
    def __init__(self, org: _Optional[_Union[_org_pb2.Org, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class AddMemberRequest(_message.Message):
    __slots__ = ["user_id", "org_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    org_id: str
    def __init__(self, user_id: _Optional[str] = ..., org_id: _Optional[str] = ...) -> None: ...

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

class ListMembersRequest(_message.Message):
    __slots__ = ["org_id"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    def __init__(self, org_id: _Optional[str] = ...) -> None: ...

class ListMembersResponse(_message.Message):
    __slots__ = ["user_ids"]
    USER_IDS_FIELD_NUMBER: _ClassVar[int]
    user_ids: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, user_ids: _Optional[_Iterable[str]] = ...) -> None: ...

class IsMemberRequest(_message.Message):
    __slots__ = ["user_id", "org_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    org_id: str
    def __init__(self, user_id: _Optional[str] = ..., org_id: _Optional[str] = ...) -> None: ...

class IsMemberResponse(_message.Message):
    __slots__ = ["is_member"]
    IS_MEMBER_FIELD_NUMBER: _ClassVar[int]
    is_member: bool
    def __init__(self, is_member: bool = ...) -> None: ...

class GetOrgsForUserRequest(_message.Message):
    __slots__ = ["user_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class GetOrgsForUserResponse(_message.Message):
    __slots__ = ["orgs"]
    ORGS_FIELD_NUMBER: _ClassVar[int]
    orgs: _containers.RepeatedCompositeFieldContainer[_org_pb2.Org]
    def __init__(self, orgs: _Optional[_Iterable[_Union[_org_pb2.Org, _Mapping]]] = ...) -> None: ...
