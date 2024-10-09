from autokitteh_pb.deployments.v1 import deployment_pb2 as _deployment_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["deployment"]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    deployment: _deployment_pb2.Deployment
    def __init__(self, deployment: _Optional[_Union[_deployment_pb2.Deployment, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["deployment_id"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    def __init__(self, deployment_id: _Optional[str] = ...) -> None: ...

class ActivateRequest(_message.Message):
    __slots__ = ["deployment_id"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    def __init__(self, deployment_id: _Optional[str] = ...) -> None: ...

class ActivateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeactivateRequest(_message.Message):
    __slots__ = ["deployment_id"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    def __init__(self, deployment_id: _Optional[str] = ...) -> None: ...

class DeactivateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class TestRequest(_message.Message):
    __slots__ = ["deployment_id"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    def __init__(self, deployment_id: _Optional[str] = ...) -> None: ...

class TestResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["env_id", "build_id", "state", "limit", "include_session_stats"]
    ENV_ID_FIELD_NUMBER: _ClassVar[int]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_SESSION_STATS_FIELD_NUMBER: _ClassVar[int]
    env_id: str
    build_id: str
    state: _deployment_pb2.DeploymentState
    limit: int
    include_session_stats: bool
    def __init__(self, env_id: _Optional[str] = ..., build_id: _Optional[str] = ..., state: _Optional[_Union[_deployment_pb2.DeploymentState, str]] = ..., limit: _Optional[int] = ..., include_session_stats: bool = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["deployments"]
    DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
    deployments: _containers.RepeatedCompositeFieldContainer[_deployment_pb2.Deployment]
    def __init__(self, deployments: _Optional[_Iterable[_Union[_deployment_pb2.Deployment, _Mapping]]] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["deployment_id"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    def __init__(self, deployment_id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["deployment"]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    deployment: _deployment_pb2.Deployment
    def __init__(self, deployment: _Optional[_Union[_deployment_pb2.Deployment, _Mapping]] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["deployment_id"]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    deployment_id: str
    def __init__(self, deployment_id: _Optional[str] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
