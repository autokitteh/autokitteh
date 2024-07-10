import json
import types
from os import environ
from subprocess import run
from sys import executable
from tempfile import NamedTemporaryFile

from conftest import test_dir

import ak_runner.__main__ as main

# Set PYTHONPATH so `python -m ak_runner` will work
env = environ.copy()
pypath = env.get("PYTHONPATH")
pypath = f"{test_dir}:{pypath}" if pypath else str(test_dir)
env["PYTHONPATH"] = pypath


def test_help():
    cmd = [executable, "-m", "ak_runner", "-h"]
    out = run(cmd, env=env)
    assert out.returncode == 0


cls_code = """
class A: pass

def fn(): pass
"""


def test_class(tmp_path):
    with open(tmp_path / "cls.py", "w") as out:
        out.write(cls_code)

    outfile = NamedTemporaryFile(delete=False)
    outfile.close()

    cmd = [executable, "-m", "ak_runner", "inspect", str(tmp_path), outfile.name]
    out = run(cmd)
    assert out.returncode == 0

    with open(outfile.name) as fp:
        reply = json.load(fp)

    expected = [
        {"name": "A", "file": "cls.py", "line": 2},
        {"name": "fn", "file": "cls.py", "line": 4},
    ]
    assert reply == expected


def test_module_entries():
    mod = types.ModuleType("test_module")
    names = ["a", "b"]
    for name in names:
        setattr(mod, name, lambda: None)
    setattr(mod, "c", 7)  # Not a callable

    entries = main.module_entries(mod)
    assert names == sorted(entries)
