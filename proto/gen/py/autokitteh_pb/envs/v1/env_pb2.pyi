from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Env(_message.Message):
    __slots__ = ["env_id", "project_id", "name"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    project_id: str
    name: str
    def __init__(self, env_id: _Optional[str] = ..., project_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...
