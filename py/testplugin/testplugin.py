from autokitteh.api import PluginID, Value
import autokitteh.plugin


def _talk(*args: str, name: str) -> str:
    name = (name + ": ") if name else ""

    return f'{name}{" ".join(args)}'


Plugin = autokitteh.plugin.Plugin(
    id=PluginID("default.test"),
    doc="test plugin",
    members={
        'cat': 'meow',
        'dog': 'woof',
        'talk': _talk,
    },
)

__all__ = ['Plugin']
