from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Org(_message.Message):
    __slots__ = ["org_id", "name"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    name: str
    def __init__(self, org_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...
