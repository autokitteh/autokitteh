"""AutoKitteh values

Wraps and unwraps autokitteh values.
"""

import pb.autokitteh.values.v1.values_pb2 as pb
from collections import namedtuple
from datetime import datetime, timedelta, UTC

from google.protobuf.duration_pb2 import Duration
from google.protobuf.timestamp_pb2 import Timestamp


def wrap(v: any) -> pb.Value:
    """Wrap a python value into an autokitteh value.

    Tuples are considered as lists.
    Classes with __slots__ or __dict__ are wrapped as structs.
    Namedtuples are wrapped as structs.
    """

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
                    ctor=wrap(v.__class__.__name__),
                    fields={f: wrap(getattr(v, f)) for f in v._fields},
                )
            )
        return pb.Value(list=pb.List(vs=[wrap(x) for x in v]))
    if isinstance(v, list):
        return pb.Value(list=pb.List(vs=[wrap(x) for x in v]))
    if isinstance(v, set):
        return pb.Value(set=pb.Set(vs=[wrap(x) for x in v]))
    if isinstance(v, dict):
        return pb.Value(
            dict=pb.Dict(
                items=[pb.Dict.Item(k=wrap(k), v=wrap(v)) for k, v in v.items()]
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

    if hasattr(v, "__dict__"):
        return pb.Value(
            struct=pb.Struct(
                ctor=wrap(v.__class__.__name__),
                fields={k: wrap(v) for k, v in v.__dict__.items()},
            )
        )

    if hasattr(v, "__slots__"):
        return pb.Value(
            struct=pb.Struct(
                ctor=wrap(v.__class__.__name__),
                fields={k: wrap(getattr(v, k)) for k in v.__slots__},
            )
        )

    raise TypeError(f"unsupported type: {type(v)}")


def unwrap(v: pb.Value) -> any:
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
        return [unwrap(x) for x in v.list.vs]
    if v.HasField("set"):
        return set([unwrap(x) for x in v.set.vs])
    if v.HasField("dict"):
        return {unwrap(x.k): unwrap(x.v) for x in v.dict.items}
    if v.HasField("bytes"):
        return v.bytes.v
    if v.HasField("struct"):
        tpl = namedtuple(str(unwrap(v.struct.ctor)), v.struct.fields.keys())
        return tpl(*[unwrap(x) for x in v.struct.fields.values()])
    if v.HasField("time"):
        return v.time.v.ToDatetime(UTC)
    if v.HasField("duration"):
        return v.duration.v.ToTimedelta()

    raise TypeError(f"unsupported type: {v}")
