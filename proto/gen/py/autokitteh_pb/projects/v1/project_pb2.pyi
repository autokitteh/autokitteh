from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Project(_message.Message):
    __slots__ = ["project_id", "name", "owner_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    name: str
    owner_id: str
    def __init__(self, project_id: _Optional[str] = ..., name: _Optional[str] = ..., owner_id: _Optional[str] = ...) -> None: ...
