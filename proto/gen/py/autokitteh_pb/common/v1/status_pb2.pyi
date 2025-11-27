from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Status(_message.Message):
    __slots__ = ["code", "message", "fix_action"]
    class Code(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        CODE_UNSPECIFIED: _ClassVar[Status.Code]
        CODE_OK: _ClassVar[Status.Code]
        CODE_WARNING: _ClassVar[Status.Code]
        CODE_ERROR: _ClassVar[Status.Code]
    CODE_UNSPECIFIED: Status.Code
    CODE_OK: Status.Code
    CODE_WARNING: Status.Code
    CODE_ERROR: Status.Code
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    FIX_ACTION_FIELD_NUMBER: _ClassVar[int]
    code: Status.Code
    message: str
    fix_action: str
    def __init__(self, code: _Optional[_Union[Status.Code, str]] = ..., message: _Optional[str] = ..., fix_action: _Optional[str] = ...) -> None: ...
