from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from autokitteh_pb.projects.v1 import project_pb2 as _project_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["project"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    project: _project_pb2.Project
    def __init__(self, project: _Optional[_Union[_project_pb2.Project, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["project_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    def __init__(self, project_id: _Optional[str] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["project_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    def __init__(self, project_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["project_id", "name", "org_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    name: str
    org_id: str
    def __init__(self, project_id: _Optional[str] = ..., name: _Optional[str] = ..., org_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["project"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    project: _project_pb2.Project
    def __init__(self, project: _Optional[_Union[_project_pb2.Project, _Mapping]] = ...) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["project"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    project: _project_pb2.Project
    def __init__(self, project: _Optional[_Union[_project_pb2.Project, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["org_id"]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    org_id: str
    def __init__(self, org_id: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["projects"]
    PROJECTS_FIELD_NUMBER: _ClassVar[int]
    projects: _containers.RepeatedCompositeFieldContainer[_project_pb2.Project]
    def __init__(self, projects: _Optional[_Iterable[_Union[_project_pb2.Project, _Mapping]]] = ...) -> None: ...

class BuildRequest(_message.Message):
    __slots__ = ["project_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    def __init__(self, project_id: _Optional[str] = ...) -> None: ...

class BuildResponse(_message.Message):
    __slots__ = ["build_id", "error"]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    build_id: str
    error: _program_pb2.Error
    def __init__(self, build_id: _Optional[str] = ..., error: _Optional[_Union[_program_pb2.Error, _Mapping]] = ...) -> None: ...

class SetResourcesRequest(_message.Message):
    __slots__ = ["project_id", "resources"]
    class ResourcesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    resources: _containers.ScalarMap[str, bytes]
    def __init__(self, project_id: _Optional[str] = ..., resources: _Optional[_Mapping[str, bytes]] = ...) -> None: ...

class SetResourcesResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DownloadResourcesRequest(_message.Message):
    __slots__ = ["project_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    def __init__(self, project_id: _Optional[str] = ...) -> None: ...

class DownloadResourcesResponse(_message.Message):
    __slots__ = ["resources"]
    class ResourcesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    resources: _containers.ScalarMap[str, bytes]
    def __init__(self, resources: _Optional[_Mapping[str, bytes]] = ...) -> None: ...

class ExportRequest(_message.Message):
    __slots__ = ["project_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    def __init__(self, project_id: _Optional[str] = ...) -> None: ...

class ExportResponse(_message.Message):
    __slots__ = ["zip_archive"]
    ZIP_ARCHIVE_FIELD_NUMBER: _ClassVar[int]
    zip_archive: bytes
    def __init__(self, zip_archive: _Optional[bytes] = ...) -> None: ...

class LintRequest(_message.Message):
    __slots__ = ["project_id", "resources", "manifest_file"]
    class ResourcesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    MANIFEST_FILE_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    resources: _containers.ScalarMap[str, bytes]
    manifest_file: str
    def __init__(self, project_id: _Optional[str] = ..., resources: _Optional[_Mapping[str, bytes]] = ..., manifest_file: _Optional[str] = ...) -> None: ...

class CheckViolation(_message.Message):
    __slots__ = ["location", "level", "message", "rule_id"]
    class Level(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        LEVEL_UNSPECIFIED: _ClassVar[CheckViolation.Level]
        LEVEL_WARNING: _ClassVar[CheckViolation.Level]
        LEVEL_ERROR: _ClassVar[CheckViolation.Level]
    LEVEL_UNSPECIFIED: CheckViolation.Level
    LEVEL_WARNING: CheckViolation.Level
    LEVEL_ERROR: CheckViolation.Level
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    LEVEL_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    RULE_ID_FIELD_NUMBER: _ClassVar[int]
    location: _program_pb2.CodeLocation
    level: CheckViolation.Level
    message: str
    rule_id: str
    def __init__(self, location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., level: _Optional[_Union[CheckViolation.Level, str]] = ..., message: _Optional[str] = ..., rule_id: _Optional[str] = ...) -> None: ...

class LintResponse(_message.Message):
    __slots__ = ["violations"]
    VIOLATIONS_FIELD_NUMBER: _ClassVar[int]
    violations: _containers.RepeatedCompositeFieldContainer[CheckViolation]
    def __init__(self, violations: _Optional[_Iterable[_Union[CheckViolation, _Mapping]]] = ...) -> None: ...
