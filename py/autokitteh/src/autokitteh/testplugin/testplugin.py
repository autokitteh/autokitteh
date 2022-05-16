from typing import Any

from autokitteh.api import PluginID, Value, FuncToValue
from autokitteh.plugin import Plugin

import autokitteh.proto.values.values_pb2 as values


def foo(_name: str) -> str:
    return _name


def sumf(*args: Any) -> int:
    return sum(args)


def dictf(**kwargs: Value) -> dict[str, Any]:
    return kwargs


def cat(name: str, **kwargs: Any) -> Any:
    vocalization = kwargs['voc']

    def kitten(name: Any, **kwargs: Any) -> str:
        return f'{name}: {vocalization}'

    return kitten


TestPlugin = Plugin(
    id=PluginID('default.test'),
    doc='test plugin',
    members={
        'cat_sound': Value.wrap('meow'),
        'dog_sound': 'woof',
        'foo': foo,
        'sum': sumf,
        'dict': dictf,
        'cat': cat,
    },
)
