import ast
import json
import pickle
import sys
import types
from pathlib import Path
from socket import socket, socketpair
from subprocess import run
from unittest.mock import MagicMock

import pytest

import ak_runner

test_dir = Path(__file__).absolute().parent


def test_load_code():
    calls = []

    mod_name = 'mod'
    class MockCall(ak_runner.AKCall):
        def __call__(self, fn, *args, **kw):
            if self.ignore(fn):
                return fn(*args, **kw)

            if fn.__module__ != mod_name:
                calls.append((fn, args, kw))
            return fn(*args, **kw)
    ak_call = MockCall(mod_name, ak_runner.Comm(socket()))

    mod = ak_runner.load_code('testdata', ak_call, mod_name)
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
    setattr(mod, 'c', 7)  # Not a callable

    entries = ak_runner.module_entries(mod)
    assert names == sorted(entries)

# Used by test_nested, must be global to be pickled.
def ak(fn): pass

val = 7
def outer():
    return ak(inner)

def inner():
    return val


def test_nested():
    global ak

    comm = MagicMock()
    comm.recv.side_effect = [
        {
            'type': ak_runner.MessageType.callback,
        },
        {
            'type': ak_runner.MessageType.response,
            'payload': {
                'value': pickle.dumps(val),
            }
        }
    ]
    comm.extract_activity.side_effect = [
        {
            'name': 'outer',
            'args': [],
            'kw': {},
            'data': (outer, (), {}),
        },
    ]

    ak = ak_runner.AKCall('mod1', comm)
    ak(outer)

    comm.send_activity.assert_called_once()


def sub(a, b, *, verbose=False):
    if verbose:
        print(f'{a} - {b}')
    return a - b


def test_comm():
    go, py = socketpair()

    # Callback
    comm = ak_runner.Comm(py)
    args, kw = (1, 7), {'verbose': False}
    comm.send_activity(sub, args, kw)
    data = go.recv(2048)
    assert data, 'no data'

    go.sendall(data)
    message = comm.recv(ak_runner.MessageType.callback)
    payload = comm.extract_activity(message)
    assert payload['name'] == sub.__name__
    assert payload['args'] == [str(v) for v in args]
    assert payload['kw'] == {k: str(v) for k, v in kw.items()}
    fn, args, kw = payload['data']
    assert fn == sub
    assert args == args
    assert kw == kw

    # Module
    names = ['a', 'c', 'f']
    comm.send_exported(names)
    data = go.recv(2048)
    assert data, 'no data'
    message = json.loads(data)
    assert message['type'] == ak_runner.MessageType.module
    assert message['payload']['entries'] == names


    # Done
    comm.send_done()
    data = go.recv(2048)
    assert data, 'no data'
    message = json.loads(data)
    assert message['type'] == ak_runner.MessageType.done


def test_load_simple():
    root_path = str(test_dir / 'testdata/simple')

    def action_fn(*args, **kw):
        pass

    ak_runner.load_code(root_path, action_fn, 'simple')


def in_act_2(v):
    print(f'in_act_2: {v}')

def in_act_1(v):
    print('in_act_1: in')
    in_act_2(v)
    print('in_act_2: in')


def test_in_activity():
    class Comm:
        def __init__(self):
            self.values = []
            self.num_activities = 0
            self.n = 0

        def send_activity(self, func, args, kw):
            self.num_activities += 1
            self.message = {'data': (func, args, kw)}

        def send_response(self, value):
            self.values.append(value)

        def extract_response(self, msg):
            return msg['payload']['value']

        def recv(self, *types):
            self.n += 1

            if self.n == 1:
                return {
                    'type': ak_runner.MessageType.callback,
                    'payload': self.message,
                }

            return {
                'type': ak_runner.MessageType.response,
                'payload': {'value': pickle.dumps(self.values[0])},
            }
        
        def extract_activity(self, msg):
            return msg['payload']


    comm = Comm()
    ak = ak_runner.AKCall('meow', comm)
    ak(in_act_1, 7)
    assert comm.num_activities == 1

    ak(in_act_1, 6)
    assert comm.num_activities == 2
    

name_of_cases = [
    # code, name
    ('print(1)', 'print'),
    ('requests.get("https://go.dev")', 'requests.get'),
    ('sheets.values().get("A1:B4").execute()', 'sheets.values.get.execute'),
]


@pytest.mark.parametrize('code, name', name_of_cases)
def test_name_of(code, name):
    mod = ast.parse(code)
    node = mod.body[0].value
    assert ak_runner.name_of(node) == name


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
    trans = ak_runner.Transformer('<stdin>')
    out = trans.visit(mod)
    assert transformed, ast.unparse(out)


def test_activity():
    mod_name = 'activity'
    mod = ak_runner.load_code('testdata', lambda f: f, mod_name)
    fn = mod.phone_home

    cb = ak_runner.AKCall(mod_name, None)
    assert not cb.ignore(fn)
