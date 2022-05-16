from typing import Any, NamedTuple

from autokitteh.api import PluginID, PluginDesc, PluginMemberDesc


class Plugin(NamedTuple):
    id: PluginID
    doc: str
    members: dict[str, Any]

    @property
    def desc(self) -> PluginDesc:
        return PluginDesc(
            self.id,
            self.doc,
            [_to_desc_member(k, v) for k, v in self.members.items()],
        )


def _to_desc_member(name: str, m: Any) -> PluginMemberDesc:
    return PluginMemberDesc(name=name, doc=m.__doc__ if hasattr(m, '__doc__') else '')
