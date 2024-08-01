from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Var(_message.Message):
    __slots__ = ["scope_id", "name", "value", "is_secret", "is_optional"]
    SCOPE_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    IS_SECRET_FIELD_NUMBER: _ClassVar[int]
    IS_OPTIONAL_FIELD_NUMBER: _ClassVar[int]
    scope_id: str
    name: str
    value: str
    is_secret: bool
    is_optional: bool
    def __init__(self, scope_id: _Optional[str] = ..., name: _Optional[str] = ..., value: _Optional[str] = ..., is_secret: bool = ..., is_optional: bool = ...) -> None: ...
