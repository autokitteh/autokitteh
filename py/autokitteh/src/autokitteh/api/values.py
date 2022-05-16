from typing import Any, Callable, Union, NamedTuple
import datetime
import inspect

import google.protobuf.timestamp_pb2 as timestamp_pb2

import autokitteh.proto.values.values_pb2 as values

from .exceptions import AutoKittehException


class UnhandledTypeError(AutoKittehException):
    def __init__(self, t: type) -> None:
        self._t = t
        super().__init__(f'unhandled type {t}')


class CannotUnwrapException(AutoKittehException):
    def __init__(self, t: type) -> None:
        self._t = t
        super().__init__(f'unwrappable {t}')


class InvalidValuePayloadError(AutoKittehException):
    def __init__(self, t: type) -> None:
        self._t = t
        super().__init__(f'invalid value with payload {t}')


class NoFuncToValueException(AutoKittehException):
    def __init__(self) -> None:
        super().__init__("func to value not supported")


Func = Callable[
    [
        str, # name
        list['Value'],
        dict[str, 'Value'],
        Callable[ # FuncToValue
            # (avoid cyclic type, hurts mypy https://github.com/python/mypy/issues/731)
            ...,
            'Value'
        ]
    ],
    Any
]

FuncToValue = Callable[[str, Func], 'Value']


def _no_func_to_value(name: str, func_to_value: Func) -> 'Value':
    raise NoFuncToValueException()


# If x is a naked protobuf, might throw CannotUnwrapException if it is
# unwrappable (like call or function).
def unwrap(x: Union[values.Value, 'Value']) -> Any:
    if type(x) == Value:
        try:
            return unwrap(x.pb)
        except CannotUnwrapException:
            # This is raised when underlying value is unwrappable
            # (like call or function). Better just pass the wrapping value
            # back to the caller as is. This is not necessarily an error.
            return x

    assert type(x) == values.Value

    """Return underlying value. Raises KeyError if not unwrappable."""
    unwrappers = {
        'none': None,
        'integer': x.integer.v,
        'string': x.string.v,
        'boolean': x.boolean.v,
        'float': x.float.v,
        'bytes': x.bytes.v,
        'list': [unwrap(v) for v in x.list.vs],
        'set': {unwrap(v) for v in x.set.vs},
        'dict': {
            unwrap(i.k): unwrap(i.v) for i in x.dict.items
        },
    }

    # This might throw CannotUnwrapException, which is fine. This lets the
    # caller know that we don't know how to unwrap it. This happens only when
    # an actual protobuf is passed here without the value boxing.
    try:
        return unwrappers[x.WhichOneof('type')]
    except KeyError:
        raise CannotUnwrapException(type(x))


class Value(NamedTuple):
    pb: values.Value

    @staticmethod
    def init(**kwargs: dict[str, Any]) -> 'Value':
        return Value(pb=values.Value(**kwargs))

    @property
    def unwrapped(self) -> Any:
        return unwrap(self)

    @staticmethod
    def wrap(x: Any, func_to_value: FuncToValue = _no_func_to_value) -> 'Value':
        if x is None:
            return Value(values.Value(none=getattr(values, 'None')()))

        if isinstance(x, Value):
            return x

        if isinstance(x, tuple) and hasattr(x, '_asdict'):
            # namedtuple
            return Value(values.Value(struct=values.Struct(
                ctor=values.Value(symbol=values.Symbol(name=x.__class__.__name__)),
                fields={k: Value.wrap(v, func_to_value).pb for k, v in x._asdict().items()}, # type: ignore
            )))

        if callable(x):
            return func_to_value(x.__name__, x)

        t = type(x)

        if t == int:
            v = values.Value(integer=values.Integer(v=x))
        elif t == str:
            v = values.Value(string=values.String(v=x))
        elif t == bool:
            v = values.Value(boolean=values.Boolean(v=x))
        elif t == float:
            v = values.Value(float=values.Float(v=x))
        elif t == bytes:
            v = values.Value(bytes=values.Bytes(v=x))
        elif t == list:
            v = values.Value(list=values.List(vs=[Value.wrap(xx, func_to_value).pb for xx in x]))
        elif t == set:
            v = values.Value(set=values.Set(vs=[Value.wrap(xx, func_to_value).pb for xx in x]))
        elif t == dict:
            v = values.Value(dict=values.Dict(
                items=[
                    values.DictItem(
                        k=Value.wrap(k, func_to_value).pb,
                        v=Value.wrap(v, func_to_value).pb,
                    ) for k, v in x.items()
                ]
            ))
        elif t == datetime.datetime:
            v = values.Value(time=values.Time(
                t=timestamp_pb2.Timestamp(seconds=int(x.timestamp()), nanos=0)
            ))
        else:
            raise UnhandledTypeError(t)

        return Value(v)
