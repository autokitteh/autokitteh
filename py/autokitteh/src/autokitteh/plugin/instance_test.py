from autokitteh.api import PluginID, Value, FuncToValue
from autokitteh.plugin import Plugin

import autokitteh.proto.values.values_pb2 as values
from autokitteh.testplugin import TestPlugin

from .instance import PluginInstance


def test_values() -> None:
    inst = PluginInstance(TestPlugin)

    assert inst.get_value('cat_sound') == Value.wrap('meow')
    assert inst.get_value('dog_sound') == Value.wrap('woof')


def test_simple_calls() -> None:
    inst = PluginInstance(TestPlugin)

    foov = inst.get_value('foo')

    assert foov == Value.init(call=values.Call(
        id='0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33',
        name='foo',
        issuer='default.test',
        flags=None,
    ))

    assert inst.call_value(foov, [], {}) == Value.wrap('foo')

    sumv = inst.get_value('sum')
    assert sumv

    assert inst.call_value(
        sumv,
        [Value.wrap(x) for x in [1, 2, 3]],
        {},
    ) == Value.wrap(6)

    dictf = inst.get_value('dict')
    assert dictf

    assert inst.call_value(
        dictf,
        [],
        {'cat': Value.wrap('meow'), 'dog': Value.wrap('woof')},
    ) == Value.wrap({'cat': 'meow', 'dog': 'woof'})


def test_nested_calls() -> None:
    inst = PluginInstance(TestPlugin)

    catf = inst.get_value('cat')
    assert catf

    kittenf = inst.call_value(
        catf,
        [Value.wrap('meow')],
        {'voc': Value.wrap('mew')},
    )

    assert inst.call_value(
        kittenf,
        [],
        {'name': Value.wrap('pepurr')},
    ) == Value.wrap('pepurr: mew')
