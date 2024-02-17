from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Connection(_message.Message):
    __slots__ = ["connection_id", "integration_id", "integration_token", "project_id", "name"]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATION_TOKEN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    integration_id: str
    integration_token: str
    project_id: str
    name: str
    def __init__(self, connection_id: _Optional[str] = ..., integration_id: _Optional[str] = ..., integration_token: _Optional[str] = ..., project_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...
