"""AutoKitteh values

Wraps and unwraps autokitteh values.
"""

from collections import namedtuple
from datetime import UTC, datetime, timedelta
from typing import Any, Callable

import requests

import pb.autokitteh.values.v1.values_pb2 as pb
from google.protobuf.duration_pb2 import Duration
from google.protobuf.timestamp_pb2 import Timestamp


def wrap_unhandled(v: Any) -> pb.Value:
    return pb.Value(
        struct=pb.Struct(
            ctor=wrap("__unhandled_type"),
            fields={
                "repr": wrap(repr(v)),
                "type": wrap(str(type(v))),
            },
        )
    )


def safe_wrap(v):
    """Same as wrap, but does not raise TypeError if a type is not supported.

    Instead, it returns a struct with the type and the string representation of the value.
    """

    return wrap(v, unhandled=wrap_unhandled)


def wrap(v: Any, unhandled: Callable[[Any], pb.Value] = None, history=None) -> pb.Value:
    """Wrap a python value into an autokitteh value.

    If a type is not supported, the unhandled function is called with the value if supplied.
    If not supplied, TypeError is raised.

    Tuples are considered as lists.
    Classes with __slots__ or __dict__ are wrapped as structs.
    Namedtuples are wrapped as structs.
    """

    history = history or []

    # Check for recursion.
    for vv in history:
        # Must use `is` here as we check if the same object.
        # If we use `v in history`, we would check if the value is equal which
        # might invoke `__eq__` which is not what we want.
        if v is vv:
            return pb.Value(
                struct=pb.Struct(
                    ctor=wrap("__recursive_value"),
                    fields={"value": wrap(repr(v))},
                )
            )

    # automatically pass on recursive parameters.
    def dive(vv):
        return wrap(vv, unhandled, history + [v])

    if v is None:
        return pb.Value(nothing=pb.Nothing())
    if isinstance(
        v, bool
    ):  # must be checked before int, as isinstance(True, int) == True.
        return pb.Value(boolean=pb.Boolean(v=v))
    if isinstance(v, int):
        return pb.Value(integer=pb.Integer(v=v))
    if isinstance(v, float):
        return pb.Value(float=pb.Float(v=v))
    if isinstance(v, str):
        return pb.Value(string=pb.String(v=v))
    if isinstance(v, tuple):
        if hasattr(type(v), "_fields"):  # namedtuple
            return pb.Value(
                struct=pb.Struct(
                    ctor=dive(v.__class__.__name__),
                    fields={f: dive(getattr(v, f)) for f in v._fields},
                )
            )
        return pb.Value(list=pb.List(vs=[dive(x) for x in v]))
    if isinstance(v, list):
        return pb.Value(list=pb.List(vs=[dive(x) for x in v]))
    if isinstance(v, set):
        return pb.Value(set=pb.Set(vs=[dive(x) for x in v]))
    if isinstance(v, dict):
        return pb.Value(
            dict=pb.Dict(
                items=[pb.Dict.Item(k=dive(k), v=dive(v)) for k, v in v.items()]
            )
        )
    if isinstance(v, bytes):
        return pb.Value(bytes=pb.Bytes(v=v))
    if isinstance(v, datetime):
        v = v.astimezone(UTC)
        ts = Timestamp()
        ts.FromDatetime(v)
        return pb.Value(time=pb.Time(v=ts))
    if isinstance(v, timedelta):
        d = Duration()
        d.FromTimedelta(v)
        return pb.Value(duration=pb.Duration(v=d))

    if isinstance(v, requests.Response):
        json = text = None

        max_size = 100 * 1024  # 100K
        if len(v.content) > max_size:
            json = "<Response content too large to be included>"
            text = dive(v.text[:max_size] + "...")
        else:
            text = dive(v.text)

            try:
                json = dive(v.json())
            except requests.exceptions.JSONDecodeError:
                json = pb.Value(nothing=pb.Nothing())

        return pb.Value(
            struct=pb.Struct(
                ctor=dive(v.__class__.__name__),
                fields={
                    "status_code": dive(v.status_code),
                    "headers": dive(v.headers),
                    "text": text,
                    "json": json,
                },
            )
        )

    if hasattr(v, "__dict__"):
        return pb.Value(
            struct=pb.Struct(
                ctor=dive(v.__class__.__name__),
                fields={
                    k: dive(v) for k, v in v.__dict__.items() if not k.startswith("_")
                },
            )
        )

    if hasattr(v, "__slots__"):
        return pb.Value(
            struct=pb.Struct(
                ctor=dive(v.__class__.__name__),
                fields={
                    k: dive(getattr(v, k)) for k in v.__slots__ if not k.startswith("_")
                },
            )
        )

    if unhandled:
        return unhandled(v) or pb.Value(nothing=pb.Nothing())

    raise TypeError(f"unsupported type: {type(v)}")


def unwrap(v: pb.Value, custom: Callable[[pb.Value], Any] = None) -> Any:
    """Unwrap an autokitteh value into a python value.

    Note that wrap and unwrap are guaranteed to be symmetric.
    Two notable examples:

    >>> unwrap(wrap((1, 2))) == [1, 2]
    True
    >>> class C:
    ...   def __init__(self):
    ...     self.x = 42
    >>> unwrap(wrap(C())) == namedtuple("C", {"x"})(42)
    True
    """

    if v.HasField("nothing"):
        return None
    if v.HasField("integer"):
        return v.integer.v
    if v.HasField("float"):
        return v.float.v
    if v.HasField("string"):
        return v.string.v
    if v.HasField("boolean"):
        return v.boolean.v
    if v.HasField("list"):
        return [unwrap(x, custom) for x in v.list.vs]
    if v.HasField("set"):
        return set([unwrap(x, custom) for x in v.set.vs])
    if v.HasField("dict"):
        return {unwrap(x.k, custom): unwrap(x.v, custom) for x in v.dict.items}
    if v.HasField("bytes"):
        return v.bytes.v
    if v.HasField("struct"):
        tpl = namedtuple(str(unwrap(v.struct.ctor, custom)), v.struct.fields.keys())
        return tpl(*[unwrap(x, custom) for x in v.struct.fields.values()])
    if v.HasField("time"):
        return v.time.v.ToDatetime(UTC)
    if v.HasField("duration"):
        return v.duration.v.ToTimedelta()
    if v.HasField("function"):
        return v.function
    if v.HasField("custom"):
        return unwrap(v.custom.value)

    raise TypeError(f"unsupported type: {v}")
