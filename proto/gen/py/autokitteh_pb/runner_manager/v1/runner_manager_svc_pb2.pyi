from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ContainerConfig(_message.Message):
    __slots__ = ["image"]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    image: str
    def __init__(self, image: _Optional[str] = ...) -> None: ...

class UserCode(_message.Message):
    __slots__ = ["build_artifact"]
    BUILD_ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    build_artifact: bytes
    def __init__(self, build_artifact: _Optional[bytes] = ...) -> None: ...

class StartRunnerRequest(_message.Message):
    __slots__ = ["container_config", "user_code", "vars", "worker_address", "session_id"]
    class VarsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CONTAINER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    USER_CODE_FIELD_NUMBER: _ClassVar[int]
    VARS_FIELD_NUMBER: _ClassVar[int]
    WORKER_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    container_config: ContainerConfig
    user_code: UserCode
    vars: _containers.ScalarMap[str, str]
    worker_address: str
    session_id: str
    def __init__(self, container_config: _Optional[_Union[ContainerConfig, _Mapping]] = ..., user_code: _Optional[_Union[UserCode, _Mapping]] = ..., vars: _Optional[_Mapping[str, str]] = ..., worker_address: _Optional[str] = ..., session_id: _Optional[str] = ...) -> None: ...

class StartRunnerResponse(_message.Message):
    __slots__ = ["runner_id", "runner_token", "runner_address", "error"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    RUNNER_TOKEN_FIELD_NUMBER: _ClassVar[int]
    RUNNER_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    runner_token: str
    runner_address: str
    error: str
    def __init__(self, runner_id: _Optional[str] = ..., runner_token: _Optional[str] = ..., runner_address: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class RunnerHealthRequest(_message.Message):
    __slots__ = ["runner_id"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    def __init__(self, runner_id: _Optional[str] = ...) -> None: ...

class RunnerHealthResponse(_message.Message):
    __slots__ = ["healthy", "error"]
    HEALTHY_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    healthy: bool
    error: str
    def __init__(self, healthy: bool = ..., error: _Optional[str] = ...) -> None: ...

class StopRunnerRequest(_message.Message):
    __slots__ = ["runner_id"]
    RUNNER_ID_FIELD_NUMBER: _ClassVar[int]
    runner_id: str
    def __init__(self, runner_id: _Optional[str] = ...) -> None: ...

class StopRunnerResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class HealthRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class HealthResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: str
    def __init__(self, error: _Optional[str] = ...) -> None: ...

class CapabilitiesRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class CapabilitiesResponse(_message.Message):
    __slots__ = ["type", "max_runners", "can_build"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    MAX_RUNNERS_FIELD_NUMBER: _ClassVar[int]
    CAN_BUILD_FIELD_NUMBER: _ClassVar[int]
    type: str
    max_runners: int
    can_build: bool
    def __init__(self, type: _Optional[str] = ..., max_runners: _Optional[int] = ..., can_build: bool = ...) -> None: ...
