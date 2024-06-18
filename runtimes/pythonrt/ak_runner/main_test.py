import sys
import types
from subprocess import run

from ak_runner import __main__ as main


def test_cmdline_help():
    cmd = [sys.executable, '-m', 'ak_runner', '-h']
    out = run(cmd)
    assert out.returncode == 0


def test_module_entries():
    mod = types.ModuleType('test_module')
    names = ['a', 'b']
    for name in names:
        setattr(mod, name, lambda: None)
    setattr(mod, 'c', 7)  # Not a callable

    entries = main.module_entries(mod)
    assert names == sorted(entries)

