from autokitteh_pb.vars.v1 import var_pb2 as _var_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SetRequest(_message.Message):
    __slots__ = ["vars"]
    VARS_FIELD_NUMBER: _ClassVar[int]
    vars: _containers.RepeatedCompositeFieldContainer[_var_pb2.Var]
    def __init__(self, vars: _Optional[_Iterable[_Union[_var_pb2.Var, _Mapping]]] = ...) -> None: ...

class SetResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["scope_id", "names"]
    SCOPE_ID_FIELD_NUMBER: _ClassVar[int]
    NAMES_FIELD_NUMBER: _ClassVar[int]
    scope_id: str
    names: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, scope_id: _Optional[str] = ..., names: _Optional[_Iterable[str]] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["scope_id", "names"]
    SCOPE_ID_FIELD_NUMBER: _ClassVar[int]
    NAMES_FIELD_NUMBER: _ClassVar[int]
    scope_id: str
    names: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, scope_id: _Optional[str] = ..., names: _Optional[_Iterable[str]] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["vars"]
    VARS_FIELD_NUMBER: _ClassVar[int]
    vars: _containers.RepeatedCompositeFieldContainer[_var_pb2.Var]
    def __init__(self, vars: _Optional[_Iterable[_Union[_var_pb2.Var, _Mapping]]] = ...) -> None: ...

class FindConnectionIDsRequest(_message.Message):
    __slots__ = ["integration_id", "name", "value"]
    INTEGRATION_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    integration_id: str
    name: str
    value: str
    def __init__(self, integration_id: _Optional[str] = ..., name: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...

class FindConnectionIDsResponse(_message.Message):
    __slots__ = ["connection_ids"]
    CONNECTION_IDS_FIELD_NUMBER: _ClassVar[int]
    connection_ids: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, connection_ids: _Optional[_Iterable[str]] = ...) -> None: ...
