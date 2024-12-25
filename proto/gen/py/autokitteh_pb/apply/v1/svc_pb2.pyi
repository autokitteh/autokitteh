from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ApplyRequest(_message.Message):
    __slots__ = ["manifest", "path", "project_name", "org_id"]
    MANIFEST_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    PROJECT_NAME_FIELD_NUMBER: _ClassVar[int]
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    manifest: str
    path: str
    project_name: str
    org_id: str
    def __init__(self, manifest: _Optional[str] = ..., path: _Optional[str] = ..., project_name: _Optional[str] = ..., org_id: _Optional[str] = ...) -> None: ...

class Effect(_message.Message):
    __slots__ = ["subject_id", "type", "text"]
    SUBJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    subject_id: str
    type: str
    text: str
    def __init__(self, subject_id: _Optional[str] = ..., type: _Optional[str] = ..., text: _Optional[str] = ...) -> None: ...

class ApplyResponse(_message.Message):
    __slots__ = ["logs", "project_ids", "effects"]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    PROJECT_IDS_FIELD_NUMBER: _ClassVar[int]
    EFFECTS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedScalarFieldContainer[str]
    project_ids: _containers.RepeatedScalarFieldContainer[str]
    effects: _containers.RepeatedCompositeFieldContainer[Effect]
    def __init__(self, logs: _Optional[_Iterable[str]] = ..., project_ids: _Optional[_Iterable[str]] = ..., effects: _Optional[_Iterable[_Union[Effect, _Mapping]]] = ...) -> None: ...
