from pathlib import Path

tests_dir = Path(__file__).parent.absolute()


class _Workflows:
    def __init__(self):
        self.root_dir = tests_dir / "workflows"

    def __getattr__(self, name):
        return self.root_dir / name

    def __dir__(self):
        return [p.name for p in self.root_dir.glob("*") if p.is_dir()]


workflows = _Workflows()
