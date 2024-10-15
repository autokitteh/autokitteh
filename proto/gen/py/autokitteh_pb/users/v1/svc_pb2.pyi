from autokitteh_pb.users.v1 import user_pb2 as _user_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetRequest(_message.Message):
    __slots__ = ["user_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["user"]
    USER_FIELD_NUMBER: _ClassVar[int]
    user: _user_pb2.User
    def __init__(self, user: _Optional[_Union[_user_pb2.User, _Mapping]] = ...) -> None: ...

class CreateRequest(_message.Message):
    __slots__ = ["user"]
    USER_FIELD_NUMBER: _ClassVar[int]
    user: _user_pb2.User
    def __init__(self, user: _Optional[_Union[_user_pb2.User, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["user_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class FindByProviderRequest(_message.Message):
    __slots__ = ["provider"]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    provider: _user_pb2.UserAuthProvider
    def __init__(self, provider: _Optional[_Union[_user_pb2.UserAuthProvider, _Mapping]] = ...) -> None: ...

class FindByProviderResponse(_message.Message):
    __slots__ = ["user_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class FindByProviderOrCreateRequest(_message.Message):
    __slots__ = ["provider"]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    provider: _user_pb2.UserAuthProvider
    def __init__(self, provider: _Optional[_Union[_user_pb2.UserAuthProvider, _Mapping]] = ...) -> None: ...

class FindByProviderOrCreateResponse(_message.Message):
    __slots__ = ["user_id", "created"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    created: bool
    def __init__(self, user_id: _Optional[str] = ..., created: bool = ...) -> None: ...
