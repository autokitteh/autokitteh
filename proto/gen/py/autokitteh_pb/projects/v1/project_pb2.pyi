from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Project(_message.Message):
    __slots__ = ["project_id", "name", "resources_root_url", "resource_paths"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_ROOT_URL_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_PATHS_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    name: str
    resources_root_url: str
    resource_paths: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, project_id: _Optional[str] = ..., name: _Optional[str] = ..., resources_root_url: _Optional[str] = ..., resource_paths: _Optional[_Iterable[str]] = ...) -> None: ...
