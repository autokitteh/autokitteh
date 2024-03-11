from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Module(_message.Message):
    __slots__ = ["functions", "variables"]
    class FunctionsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Function
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Function, _Mapping]] = ...) -> None: ...
    class VariablesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Variable
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Variable, _Mapping]] = ...) -> None: ...
    FUNCTIONS_FIELD_NUMBER: _ClassVar[int]
    VARIABLES_FIELD_NUMBER: _ClassVar[int]
    functions: _containers.MessageMap[str, Function]
    variables: _containers.MessageMap[str, Variable]
    def __init__(self, functions: _Optional[_Mapping[str, Function]] = ..., variables: _Optional[_Mapping[str, Variable]] = ...) -> None: ...

class Variable(_message.Message):
    __slots__ = ["description"]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    description: str
    def __init__(self, description: _Optional[str] = ...) -> None: ...

class Function(_message.Message):
    __slots__ = ["description", "documentation_url", "input", "output", "examples", "deprecated_message"]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    DOCUMENTATION_URL_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    EXAMPLES_FIELD_NUMBER: _ClassVar[int]
    DEPRECATED_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    description: str
    documentation_url: str
    input: _containers.RepeatedCompositeFieldContainer[FunctionField]
    output: _containers.RepeatedCompositeFieldContainer[FunctionField]
    examples: _containers.RepeatedCompositeFieldContainer[Example]
    deprecated_message: str
    def __init__(self, description: _Optional[str] = ..., documentation_url: _Optional[str] = ..., input: _Optional[_Iterable[_Union[FunctionField, _Mapping]]] = ..., output: _Optional[_Iterable[_Union[FunctionField, _Mapping]]] = ..., examples: _Optional[_Iterable[_Union[Example, _Mapping]]] = ..., deprecated_message: _Optional[str] = ...) -> None: ...

class FunctionField(_message.Message):
    __slots__ = ["name", "description", "type", "optional", "default_value", "kwarg", "examples"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    OPTIONAL_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_VALUE_FIELD_NUMBER: _ClassVar[int]
    KWARG_FIELD_NUMBER: _ClassVar[int]
    EXAMPLES_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    type: str
    optional: bool
    default_value: str
    kwarg: bool
    examples: _containers.RepeatedCompositeFieldContainer[Example]
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., type: _Optional[str] = ..., optional: bool = ..., default_value: _Optional[str] = ..., kwarg: bool = ..., examples: _Optional[_Iterable[_Union[Example, _Mapping]]] = ...) -> None: ...

class Example(_message.Message):
    __slots__ = ["code", "explanation"]
    CODE_FIELD_NUMBER: _ClassVar[int]
    EXPLANATION_FIELD_NUMBER: _ClassVar[int]
    code: str
    explanation: str
    def __init__(self, code: _Optional[str] = ..., explanation: _Optional[str] = ...) -> None: ...
