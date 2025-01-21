from collections import namedtuple
from datetime import UTC, datetime, timedelta

import google.protobuf.duration_pb2 as duration_pb2
import google.protobuf.timestamp_pb2 as timestamp_pb2
import pb.autokitteh.values.v1.values_pb2 as pb
import pytest
from threading import Lock
import requests

from values import unwrap, wrap, safe_wrap, wrap_unhandled


def intv(n):
    return pb.Value(integer=pb.Integer(v=n))


wrap_test_cases = [
    pytest.param(None, pb.Value(nothing=pb.Nothing()), id="None"),
    pytest.param(True, pb.Value(boolean=pb.Boolean(v=True)), id="True"),
    pytest.param(False, pb.Value(boolean=pb.Boolean(v=False)), id="False"),
    pytest.param(1, intv(1), id="int"),
    pytest.param(1.0, pb.Value(float=pb.Float(v=1.0)), id="float"),
    pytest.param("meow", pb.Value(string=pb.String(v="meow")), id="str"),
    pytest.param(
        [1, 2, 3],
        pb.Value(
            list=pb.List(
                vs=[intv(1), intv(2), intv(3)],
            ),
        ),
        id="list",
    ),
    pytest.param(
        set([1, 2, 3]),
        pb.Value(
            set=pb.Set(vs=[intv(1), intv(2), intv(3)]),
        ),
        id="set",
    ),
    pytest.param(
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
        id="dict",
    ),
    pytest.param(
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
        id="namedtuple",
    ),
    pytest.param(
        timedelta(days=1),
        pb.Value(duration=pb.Duration(v=duration_pb2.Duration(seconds=86400))),
        id="timedelta",
    ),
    pytest.param(
        datetime.fromtimestamp(1663595726).astimezone(tz=UTC),
        pb.Value(
            time=pb.Time(
                v=timestamp_pb2.Timestamp(
                    seconds=1663595726,
                    nanos=0,
                )
            )
        ),
        id="datetime",
    ),
]


@pytest.mark.parametrize("u,v", wrap_test_cases)
def test_wrap(u, v):
    assert wrap(u) == v
    assert safe_wrap(u) == v
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


def test_safe_wrap():
    v = Lock()
    u = safe_wrap(v)
    assert u == wrap_unhandled(v)
    assert u.struct.ctor == wrap("__unhandled_type")
    assert u.struct.fields.keys() == {"type", "repr"}
    assert u.struct.fields["type"] == wrap("<class '_thread.lock'>")


def test_recursive_wrap():
    class C:
        def __init__(self):
            self.x = "meow"
            self.next = self

    c = C()

    assert wrap(c) == pb.Value(
        struct=pb.Struct(
            ctor=wrap("C"),
            fields={
                "x": wrap("meow"),
                "next": pb.Value(
                    struct=pb.Struct(
                        ctor=wrap("__recursive_value"),
                        fields={"value": wrap(repr(c))},
                    )
                ),
            },
        ),
    )
