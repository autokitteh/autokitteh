import ak_runner
import json
import sys
from subprocess import run
from pathlib import Path
import types


test_dir = Path(__file__).absolute().parent


def test_load_code():
    calls = []

    def action_fn(fn, *args, **kw):
        calls.append((fn, args, kw))
        return fn(*args, **kw)

    mod_name = 'mod'
    mod = ak_runner.load_code('testdata', action_fn, mod_name)
    fn = getattr(mod, 'parse', None)
    assert fn, 'parse not found'

    obj = {'x': 1, 'y': 2}
    out = fn(json.dumps(obj))
    assert out == obj, 'parse fail'
    assert len(calls) == 1, 'calls'
    fn = calls[0][0]
    name = fn.__module__ + '.' + fn.__name__
    assert name == 'json.loads'


def test_cmdline_help():
    py_file = str(test_dir / 'ak_runner.py')
    cmd = [sys.executable, py_file, '-h']
    out = run(cmd)
    assert out.returncode == 0


def test_module_entries():
    mod = types.ModuleType('test_module')
    names = ['a', 'b']
    for name in names:
        setattr(mod, name, lambda: None)
    mod.c = 7

    entries = ak_runner.module_entries(mod)
    assert names == sorted(entries)
