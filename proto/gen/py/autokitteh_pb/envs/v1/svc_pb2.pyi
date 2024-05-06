from autokitteh_pb.envs.v1 import env_pb2 as _env_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ListRequest(_message.Message):
    __slots__ = ["project_id"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    def __init__(self, project_id: _Optional[str] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["envs"]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    envs: _containers.RepeatedCompositeFieldContainer[_env_pb2.Env]
    def __init__(self, envs: _Optional[_Iterable[_Union[_env_pb2.Env, _Mapping]]] = ...) -> None: ...

class CreateRequest(_message.Message):
    __slots__ = ["env"]
    ENV_FIELD_NUMBER: _ClassVar[int]
    env: _env_pb2.Env
    def __init__(self, env: _Optional[_Union[_env_pb2.Env, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["env_id"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    def __init__(self, env_id: _Optional[str] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["env_id", "name", "project_id"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    name: str
    project_id: str
    def __init__(self, env_id: _Optional[str] = ..., name: _Optional[str] = ..., project_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["env"]
    ENV_FIELD_NUMBER: _ClassVar[int]
    env: _env_pb2.Env
    def __init__(self, env: _Optional[_Union[_env_pb2.Env, _Mapping]] = ...) -> None: ...

class RemoveRequest(_message.Message):
    __slots__ = ["env_id"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    def __init__(self, env_id: _Optional[str] = ...) -> None: ...

class RemoveResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["env"]
    ENV_FIELD_NUMBER: _ClassVar[int]
    env: _env_pb2.Env
    def __init__(self, env: _Optional[_Union[_env_pb2.Env, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
