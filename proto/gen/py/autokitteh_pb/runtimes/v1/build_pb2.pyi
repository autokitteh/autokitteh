from autokitteh_pb.program.v1 import program_pb2 as _program_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class BuildArtifact(_message.Message):
    __slots__ = ["requirements", "exports", "compiled_data"]
    class CompiledDataEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    REQUIREMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORTS_FIELD_NUMBER: _ClassVar[int]
    COMPILED_DATA_FIELD_NUMBER: _ClassVar[int]
    requirements: _containers.RepeatedCompositeFieldContainer[Requirement]
    exports: _containers.RepeatedCompositeFieldContainer[Export]
    compiled_data: _containers.ScalarMap[str, bytes]
    def __init__(self, requirements: _Optional[_Iterable[_Union[Requirement, _Mapping]]] = ..., exports: _Optional[_Iterable[_Union[Export, _Mapping]]] = ..., compiled_data: _Optional[_Mapping[str, bytes]] = ...) -> None: ...

class Requirement(_message.Message):
    __slots__ = ["location", "url", "symbol"]
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    location: _program_pb2.CodeLocation
    url: str
    symbol: str
    def __init__(self, location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., url: _Optional[str] = ..., symbol: _Optional[str] = ...) -> None: ...

class Export(_message.Message):
    __slots__ = ["location", "symbol"]
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    SYMBOL_FIELD_NUMBER: _ClassVar[int]
    location: _program_pb2.CodeLocation
    symbol: str
    def __init__(self, location: _Optional[_Union[_program_pb2.CodeLocation, _Mapping]] = ..., symbol: _Optional[str] = ...) -> None: ...
