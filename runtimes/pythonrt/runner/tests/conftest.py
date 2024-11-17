import sys
from pathlib import Path

tests_dir = Path(__file__).parent.absolute()


class _Workflows:
    def __init__(self):
        self.root_dir = tests_dir / "workflows"

    def __getattr__(self, name):
        return self.root_dir / name


workflows = _Workflows()


def clear_module_cache(*names):
    """If a module is already loaded, our custom loader won't be called."""
    for name in names:
        sys.modules.pop(name, 0)
