from autokitteh_pb.triggers.v1 import trigger_pb2 as _trigger_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["trigger"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    trigger: _trigger_pb2.Trigger
    def __init__(self, trigger: _Optional[_Union[_trigger_pb2.Trigger, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["trigger_id"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    trigger_id: str
    def __init__(self, trigger_id: _Optional[str] = ...) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["trigger"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    trigger: _trigger_pb2.Trigger
    def __init__(self, trigger: _Optional[_Union[_trigger_pb2.Trigger, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["trigger_id"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    trigger_id: str
    def __init__(self, trigger_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["trigger_id"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    trigger_id: str
    def __init__(self, trigger_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["trigger"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    trigger: _trigger_pb2.Trigger
    def __init__(self, trigger: _Optional[_Union[_trigger_pb2.Trigger, _Mapping]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["env_id", "connection_id", "project_id", "source_type"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    connection_id: str
    project_id: str
    source_type: _trigger_pb2.Trigger.SourceType
    def __init__(self, env_id: _Optional[str] = ..., connection_id: _Optional[str] = ..., project_id: _Optional[str] = ..., source_type: _Optional[_Union[_trigger_pb2.Trigger.SourceType, str]] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["triggers"]
    TRIGGERS_FIELD_NUMBER: _ClassVar[int]
    triggers: _containers.RepeatedCompositeFieldContainer[_trigger_pb2.Trigger]
    def __init__(self, triggers: _Optional[_Iterable[_Union[_trigger_pb2.Trigger, _Mapping]]] = ...) -> None: ...
