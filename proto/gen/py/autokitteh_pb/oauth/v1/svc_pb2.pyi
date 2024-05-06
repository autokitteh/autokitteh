from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RegisterRequest(_message.Message):
    __slots__ = ["id", "config"]
    ID_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    id: str
    config: OAuthConfig
    def __init__(self, id: _Optional[str] = ..., config: _Optional[_Union[OAuthConfig, _Mapping]] = ...) -> None: ...

class RegisterResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["config"]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: OAuthConfig
    def __init__(self, config: _Optional[_Union[OAuthConfig, _Mapping]] = ...) -> None: ...

class StartFlowRequest(_message.Message):
    __slots__ = ["id", "connection_id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    connection_id: str
    def __init__(self, id: _Optional[str] = ..., connection_id: _Optional[str] = ...) -> None: ...

class StartFlowResponse(_message.Message):
    __slots__ = ["url"]
    URL_FIELD_NUMBER: _ClassVar[int]
    url: str
    def __init__(self, url: _Optional[str] = ...) -> None: ...

class ExchangeRequest(_message.Message):
    __slots__ = ["id", "state", "code"]
    ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    id: str
    state: str
    code: str
    def __init__(self, id: _Optional[str] = ..., state: _Optional[str] = ..., code: _Optional[str] = ...) -> None: ...

class ExchangeResponse(_message.Message):
    __slots__ = ["access_token", "refresh_token", "token_type", "expiry"]
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    TOKEN_TYPE_FIELD_NUMBER: _ClassVar[int]
    EXPIRY_FIELD_NUMBER: _ClassVar[int]
    access_token: str
    refresh_token: str
    token_type: str
    expiry: int
    def __init__(self, access_token: _Optional[str] = ..., refresh_token: _Optional[str] = ..., token_type: _Optional[str] = ..., expiry: _Optional[int] = ...) -> None: ...

class OAuthConfig(_message.Message):
    __slots__ = ["client_id", "client_secret", "auth_url", "device_auth_url", "token_url", "redirect_url", "auth_style", "options", "scopes"]
    class OptionsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CLIENT_ID_FIELD_NUMBER: _ClassVar[int]
    CLIENT_SECRET_FIELD_NUMBER: _ClassVar[int]
    AUTH_URL_FIELD_NUMBER: _ClassVar[int]
    DEVICE_AUTH_URL_FIELD_NUMBER: _ClassVar[int]
    TOKEN_URL_FIELD_NUMBER: _ClassVar[int]
    REDIRECT_URL_FIELD_NUMBER: _ClassVar[int]
    AUTH_STYLE_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    SCOPES_FIELD_NUMBER: _ClassVar[int]
    client_id: str
    client_secret: str
    auth_url: str
    device_auth_url: str
    token_url: str
    redirect_url: str
    auth_style: int
    options: _containers.ScalarMap[str, str]
    scopes: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, client_id: _Optional[str] = ..., client_secret: _Optional[str] = ..., auth_url: _Optional[str] = ..., device_auth_url: _Optional[str] = ..., token_url: _Optional[str] = ..., redirect_url: _Optional[str] = ..., auth_style: _Optional[int] = ..., options: _Optional[_Mapping[str, str]] = ..., scopes: _Optional[_Iterable[str]] = ...) -> None: ...
