import pb.autokitteh.values.v1.values_pb2 as pb
from values import DefaultValueWrapper
from collections import namedtuple
from datetime import datetime, timedelta, UTC
import google.protobuf.timestamp_pb2 as timestamp_pb2
import google.protobuf.duration_pb2 as duration_pb2


w = DefaultValueWrapper


def check(u, **kwargs):
    v = pb.Value(**kwargs)
    assert w.wrap(u) == v
    assert w.unwrap(v) == u


def test_wrap():
    check(None, nothing=pb.Nothing())
    check(1, integer=pb.Integer(v=1))
    check(1.0, float=pb.Float(v=1.0))
    check("meow", string=pb.String(v="meow"))
    check(True, boolean=pb.Boolean(v=True))
    check(False, boolean=pb.Boolean(v=False))
    check(b"meow", bytes=pb.Bytes(v=b"meow"))

    lv = pb.Value(
        list=pb.List(
            vs=[
                pb.Value(integer=pb.Integer(v=1)),
                pb.Value(integer=pb.Integer(v=2)),
                pb.Value(integer=pb.Integer(v=3)),
            ],
        )
    )

    assert w.wrap((1, 2, 3)) == lv
    assert w.unwrap(lv) == [1, 2, 3]  # unwrap always return tuples as lists.

    check(
        [1, 2, 3],
        list=pb.List(
            vs=[
                pb.Value(integer=pb.Integer(v=1)),
                pb.Value(integer=pb.Integer(v=2)),
                pb.Value(integer=pb.Integer(v=3)),
            ],
        ),
    )

    check(
        set([1, 2, 3]),
        set=pb.Set(
            vs=[
                pb.Value(integer=pb.Integer(v=1)),
                pb.Value(integer=pb.Integer(v=2)),
                pb.Value(integer=pb.Integer(v=3)),
            ],
        ),
    )

    check(
        {"a": 1, "b": 2},
        dict=pb.Dict(
            items=[
                pb.Dict.Item(
                    k=pb.Value(string=pb.String(v="a")), v=pb.Value(integer=pb.Integer(v=1))
                ),
                pb.Dict.Item(
                    k=pb.Value(string=pb.String(v="b")), v=pb.Value(integer=pb.Integer(v=2))
                ),
            ],
        ),
    )

    check(
        namedtuple("Point", ["y", "x"])(2, 1),
        struct=pb.Struct(
            ctor=pb.Value(string=pb.String(v="Point")),
            fields={
                "x": pb.Value(integer=pb.Integer(v=1)),
                "y": pb.Value(integer=pb.Integer(v=2)),
            },
        ),
    )

    check(
        datetime.strptime("09/19/22 13:55:26", "%m/%d/%y %H:%M:%S").astimezone(tz=UTC),
        time=pb.Time(
            v=timestamp_pb2.Timestamp(
                seconds=1663620926,
                nanos=0,
            )
        ),
    )

    check(
        timedelta(days=1),
        duration=pb.Duration(v=duration_pb2.Duration(seconds=86400)),
    )

    def mkv(n):
        return pb.Value(
            struct=pb.Struct(
                ctor=w.wrap(n),
                fields={
                    "a": pb.Value(integer=pb.Integer(v=1)),
                    "b": pb.Value(string=pb.String(v="meow")),
                },
            )
        )

    class C:
        def __init__(self):
            self.a = 1
            self.b = "meow"

    cv = mkv("C")

    assert w.wrap(C()) == cv

    assert w.unwrap(cv) == namedtuple("C", ["a", "b"])(1, "meow")

    class D:
        __slots__ = ("a", "b")

        def __init__(self):
            self.a = 1
            self.b = "meow"

    dv = mkv("D")

    assert w.wrap(D()) == dv

    assert w.unwrap(dv) == namedtuple("D", ["a", "b"])(1, "meow")
