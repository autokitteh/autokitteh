import ast
import json
from os import environ
from pathlib import Path
from socket import socket
from unittest.mock import MagicMock

import pytest

import ak_runner
from ak_runner import loader

test_dir = Path(__file__).absolute().parent.parent
simple_dir = test_dir / 'testdata/simple'


def test_load_code():
    calls = []

    mod_name = 'mod'
    class MockCall(ak_runner.AKCall):
        def __init__(self, *args, **kw):
            super().__init__(*args, **kw)
            self.loading = False

        def __call__(self, fn, *args, **kw):
            if not self.should_run_as_activity(fn):
                return fn(*args, **kw)

            if fn.__module__ != mod_name:
                calls.append((fn, args, kw))
            return fn(*args, **kw)
    ak_call = MockCall(ak_runner.Comm(socket()))

    mod = ak_runner.load_code('testdata', ak_call, mod_name)
    ak_call.set_module(mod)
    fn = getattr(mod, 'parse', None)
    assert fn, 'parse not found'

    obj = {'x': 1, 'y': 2}
    out = fn(json.dumps(obj))
    assert out == obj, 'parse fail'
    assert len(calls) == 1, 'calls'
    fn = calls[0][0]
    name = fn.__module__ + '.' + fn.__name__
    assert name == 'json.loads'


def test_load_twice(tmp_path):
    mod_name = 'x'
    file_name = tmp_path / (mod_name + '.py')
    var, val = 'x', 1
    with open(file_name, 'w') as out:
        print(f'{var} = {val}', file=out)

    mod = ak_runner.load_code(tmp_path, lambda x: x, mod_name)
    assert getattr(mod, var) == val

    # See that module is not cached.
    val += 1
    with open(file_name, 'w') as out:
        print(f'{var} = {val}', file=out)

    mod = ak_runner.load_code(tmp_path, lambda x: x, mod_name)
    assert getattr(mod, var) == val


def test_load_simple():
    root_path = str(simple_dir)

    def action_fn(*args, **kw):
        pass

    ak_runner.load_code(root_path, action_fn, 'simple')


name_of_cases = [
    # code, name
    ('print(1)', 'print'),
    ('requests.get("https://go.dev")', 'requests.get'),
    ('sheets.values().get("A1:B4").execute()', 'sheets.values.get.execute'),
    ('label["name"].lower()', 'label["name"].lower'),
]


@pytest.mark.parametrize('code, name', name_of_cases)
def test_name_of(code, name):
    mod = ast.parse(code)
    node = mod.body[0].value
    assert loader.name_of(node) == name


transform_cases = [
    # code, transformed
    ('get(1)', '_ak_call(get, 1)'),
    ('requests.get("https://go.dev")', '_ak_call(requests.get, "https://go.dev")'),
    (
        'sheets.values().get("A1:B4").execute()', 
        '_ak_call(_ak_call(_ak_call(google.sheets.values).get, "A1:A10").execute)',
    ),
    ('add(get(1), get(2))', '_ak_call(add, _ak_call(get, 1), _ak_call(get, 2))'),
]

@pytest.mark.parametrize('code, transformed', transform_cases)
def test_transform(code, transformed):
    mod = ast.parse(code)
    trans = loader.Transformer('<stdin>')
    out = trans.visit(mod)
    assert transformed, ast.unparse(out)


def test_module_level():
    comm = MagicMock()
    root_path = str(test_dir / 'testdata')
    akc = ak_runner.AKCall(comm)
    mod = ak_runner.load_code(root_path, akc, 'modlevel')
    assert mod.home == environ['HOME']
    akc.set_module(mod)

    mod.on_event(None)
    assert mod.ncalls == 1
