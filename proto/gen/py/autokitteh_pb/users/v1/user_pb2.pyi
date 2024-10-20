from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class UserAuthProvider(_message.Message):
    __slots__ = ["name", "user_id", "email", "data"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    name: str
    user_id: str
    email: str
    data: bytes
    def __init__(self, name: _Optional[str] = ..., user_id: _Optional[str] = ..., email: _Optional[str] = ..., data: _Optional[bytes] = ...) -> None: ...

class User(_message.Message):
    __slots__ = ["user_id", "primary_email", "auth_providers"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    PRIMARY_EMAIL_FIELD_NUMBER: _ClassVar[int]
    AUTH_PROVIDERS_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    primary_email: str
    auth_providers: _containers.RepeatedCompositeFieldContainer[UserAuthProvider]
    def __init__(self, user_id: _Optional[str] = ..., primary_email: _Optional[str] = ..., auth_providers: _Optional[_Iterable[_Union[UserAuthProvider, _Mapping]]] = ...) -> None: ...
