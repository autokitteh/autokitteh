from autokitteh_pb.sessions.v1 import session_pb2 as _session_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeploymentState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    DEPLOYMENT_STATE_UNSPECIFIED: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_ACTIVE: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_TESTING: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_DRAINING: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_INACTIVE: _ClassVar[DeploymentState]
DEPLOYMENT_STATE_UNSPECIFIED: DeploymentState
DEPLOYMENT_STATE_ACTIVE: DeploymentState
DEPLOYMENT_STATE_TESTING: DeploymentState
DEPLOYMENT_STATE_DRAINING: DeploymentState
DEPLOYMENT_STATE_INACTIVE: DeploymentState

class Deployment(_message.Message):
    __slots__ = ["project_id", "deployment_id", "build_id", "state", "created_at", "updated_at", "sessions_stats"]
    class SessionStats(_message.Message):
        __slots__ = ["state", "count"]
        STATE_FIELD_NUMBER: _ClassVar[int]
        COUNT_FIELD_NUMBER: _ClassVar[int]
        state: _session_pb2.SessionStateType
        count: int
        def __init__(self, state: _Optional[_Union[_session_pb2.SessionStateType, str]] = ..., count: _Optional[int] = ...) -> None: ...
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    BUILD_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    SESSIONS_STATS_FIELD_NUMBER: _ClassVar[int]
    project_id: str
    deployment_id: str
    build_id: str
    state: DeploymentState
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    sessions_stats: _containers.RepeatedCompositeFieldContainer[Deployment.SessionStats]
    def __init__(self, project_id: _Optional[str] = ..., deployment_id: _Optional[str] = ..., build_id: _Optional[str] = ..., state: _Optional[_Union[DeploymentState, str]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., sessions_stats: _Optional[_Iterable[_Union[Deployment.SessionStats, _Mapping]]] = ...) -> None: ...
