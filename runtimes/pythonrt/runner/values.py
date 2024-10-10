"""AutoKitteh values

Wraps and unwraps autokitteh values.
"""

import pb.autokitteh.values.v1.values_pb2 as pb
from collections import namedtuple
from datetime import datetime, timedelta, UTC

from google.protobuf.duration_pb2 import Duration
from google.protobuf.timestamp_pb2 import Timestamp


class ValueWrapper:
    # TODO: This is an instance method as we might add options to the wrapper in the future.
    def wrap(self, v: any) -> pb.Value:
        """Wrap a python value into an autokitteh value.

        Tuples are considered as lists.
        Classes with __slots__ or __dict__ are wrapped as structs.
        Namedtuples are wrapped as structs.
        """

        if v is None:
            return pb.Value(nothing=pb.Nothing())
        if isinstance(v, bool):  # must be checked before int, as isinstance(True, int) == True.
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
                        ctor=self.wrap(v.__class__.__name__),
                        fields={f: self.wrap(getattr(v, f)) for f in v._fields},
                    )
                )
            return pb.Value(list=pb.List(vs=[self.wrap(x) for x in v]))
        if isinstance(v, list):
            return pb.Value(list=pb.List(vs=[self.wrap(x) for x in v]))
        if isinstance(v, set):
            return pb.Value(set=pb.Set(vs=[self.wrap(x) for x in v]))
        if isinstance(v, dict):
            return pb.Value(
                dict=pb.Dict(
                    items=[pb.Dict.Item(k=self.wrap(k), v=self.wrap(v)) for k, v in v.items()]
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
                    ctor=self.wrap(v.__class__.__name__),
                    fields={k: self.wrap(v) for k, v in v.__dict__.items()},
                )
            )

        if hasattr(v, "__slots__"):
            return pb.Value(
                struct=pb.Struct(
                    ctor=self.wrap(v.__class__.__name__),
                    fields={k: self.wrap(getattr(v, k)) for k in v.__slots__},
                )
            )

        raise ValueError(f"unsupported type: {type(v)}")

    # TODO: This is an instance method as we might add options to the wrapper in the future.
    def unwrap(self, v: pb.Value) -> any:
        """Unwrap an autokitteh value into a python value.

        Note that wrap and unwrap are guaranteed to be symmetric.
        Two notable examples:
            unwrap(wrap((1, 2))) = [1, 2]

            class C:
                def __init__(self):
                    self.x = 42

            unwrap(C()) = namedtuple("C", {"x"})(42)
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
            return [self.unwrap(x) for x in v.list.vs]
        if v.HasField("set"):
            return set([self.unwrap(x) for x in v.set.vs])
        if v.HasField("dict"):
            return {self.unwrap(x.k): self.unwrap(x.v) for x in v.dict.items}
        if v.HasField("bytes"):
            return v.bytes.v
        if v.HasField("struct"):
            tpl = namedtuple(str(self.unwrap(v.struct.ctor)), v.struct.fields.keys())
            return tpl(*[self.unwrap(x) for x in v.struct.fields.values()])
        if v.HasField("time"):
            return datetime.fromtimestamp(v.time.v.seconds, UTC)
        if v.HasField("duration"):
            return timedelta(seconds=v.duration.v.seconds, milliseconds=v.duration.v.nanos / 1000)

        raise ValueError(f"unsupported type: {v}")


DefaultValueWrapper = ValueWrapper()
