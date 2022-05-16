from collections import namedtuple
import datetime
import pytest

from .values import Value, unwrap, CannotUnwrapException
import autokitteh.proto.values.values_pb2 as values_pb


def test_wrap() -> None:
    assert Value.wrap(None).unwrapped is None

    assert Value.wrap(1).pb.integer.v == 1
    assert Value.wrap(1).unwrapped == 1

    assert Value.wrap("one").pb.string.v == "one"
    assert Value.wrap("one").unwrapped == "one"

    assert Value.wrap(1.0).pb.float.v == 1.0
    assert Value.wrap(1.0).unwrapped == 1.0

    assert not Value.wrap(False).pb.boolean.v
    assert not Value.wrap(False).unwrapped

    assert Value.wrap(True).pb.boolean.v
    assert Value.wrap(True).unwrapped

    assert Value.wrap(b'meow').pb.bytes.v == b'meow'
    assert Value.wrap(b'meow').unwrapped == b'meow'

    assert [x.integer.v for x in Value.wrap([1, 2, 3]).pb.list.vs] == [1, 2, 3]
    assert Value.wrap([1, 2, 3]).unwrapped == [1, 2, 3]

    assert [x.integer.v for x in Value.wrap({1, 2, 3}).pb.set.vs] == [1, 2, 3]
    assert Value.wrap({1, 2, 3}).unwrapped == {1, 2, 3}

    assert Value.wrap({"one": 1, "two": 2}).unwrapped == {"one": 1, "two": 2}

    assert {
        i.k.string.v: i.v.integer.v for i in Value.wrap({"one": 1, "two": 2}).pb.dict.items
    } == {"one": 1, "two": 2}
    assert Value.wrap(datetime.datetime(year=2022, month=2, day=26)).pb.time.t.seconds == 1645862400

    x = namedtuple("X", ("one", "two"))(1, 2)
    v = Value.wrap(x)

    assert v.unwrapped == v

    assert v.pb.struct.ctor.symbol.name == 'X'
    assert v.pb.struct.fields['one'].integer.v == 1
    assert v.pb.struct.fields['two'].integer.v == 2


def test_unwrap_pb() -> None:
    v = Value.init(integer=values_pb.Integer(v=42))
    assert unwrap(v) == 42


def test_unwrappable() -> None:
    v = Value.init(call=values_pb.Call())
    assert unwrap(v) == v

    with pytest.raises(CannotUnwrapException):
        unwrap(Value.init(call=values_pb.Call()).pb)
