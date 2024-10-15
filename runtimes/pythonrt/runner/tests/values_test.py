import pb.autokitteh.values.v1.values_pb2 as pb
from values import wrap, unwrap
from collections import namedtuple
from datetime import datetime, timedelta, UTC
import pytest
import google.protobuf.timestamp_pb2 as timestamp_pb2
import google.protobuf.duration_pb2 as duration_pb2


def intv(n):
    return pb.Value(integer=pb.Integer(v=n))


wrap_test_cases = [
    (None, pb.Value(nothing=pb.Nothing())),
    (True, pb.Value(boolean=pb.Boolean(v=True))),
    (False, pb.Value(boolean=pb.Boolean(v=False))),
    (1, intv(1)),
    (1.0, pb.Value(float=pb.Float(v=1.0))),
    ("meow", pb.Value(string=pb.String(v="meow"))),
    (
        [1, 2, 3],
        pb.Value(
            list=pb.List(
                vs=[intv(1), intv(2), intv(3)],
            ),
        ),
    ),
    (
        set([1, 2, 3]),
        pb.Value(
            set=pb.Set(vs=[intv(1), intv(2), intv(3)]),
        ),
    ),
    (
        {"a": 1, "b": 2},
        pb.Value(
            dict=pb.Dict(
                items=[
                    pb.Dict.Item(
                        k=pb.Value(string=pb.String(v="a")),
                        v=intv(1),
                    ),
                    pb.Dict.Item(
                        k=pb.Value(string=pb.String(v="b")),
                        v=intv(2),
                    ),
                ],
            )
        ),
    ),
    (
        namedtuple("Point", ["y", "x"])(2, 1),
        pb.Value(
            struct=pb.Struct(
                ctor=pb.Value(string=pb.String(v="Point")),
                fields={
                    "x": intv(1),
                    "y": intv(2),
                },
            )
        ),
    ),
    (
        timedelta(days=1),
        pb.Value(duration=pb.Duration(v=duration_pb2.Duration(seconds=86400))),
    ),
    (
        datetime.strptime("09/19/22 13:55:26", "%m/%d/%y %H:%M:%S").astimezone(tz=UTC),
        pb.Value(
            time=pb.Time(
                v=timestamp_pb2.Timestamp(
                    seconds=1663595726,
                    nanos=0,
                )
            )
        ),
    ),
]


@pytest.mark.parametrize("u,v", wrap_test_cases)
def test_wrap(u, v):
    assert wrap(u) == v
    assert unwrap(v) == u


def test_special_wraps():
    lv = pb.Value(
        list=pb.List(
            vs=[
                pb.Value(integer=pb.Integer(v=1)),
                pb.Value(integer=pb.Integer(v=2)),
                pb.Value(integer=pb.Integer(v=3)),
            ],
        )
    )

    assert wrap((1, 2, 3)) == lv
    assert unwrap(lv) == [1, 2, 3]  # unwrap always return tuples as lists.

    def mkv(n):
        return pb.Value(
            struct=pb.Struct(
                ctor=wrap(n),
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

    assert wrap(C()) == cv

    assert unwrap(cv) == namedtuple("C", ["a", "b"])(1, "meow")

    class D:
        __slots__ = ("a", "b")

        def __init__(self):
            self.a = 1
            self.b = "meow"

    dv = mkv("D")

    assert wrap(D()) == dv

    assert unwrap(dv) == namedtuple("D", ["a", "b"])(1, "meow")
