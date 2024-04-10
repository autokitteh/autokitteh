import json
import pickle
import sys
import types
from pathlib import Path
from socket import socketpair
from subprocess import run
from threading import Thread

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
    ak_call = MockCall(mod_name, None, None)

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


def test_nested():
    ak = ak_runner.AKCall('mod1')
    val = 7

    def outer():
        return ak(inner)

    def inner():
        return val

    thr = Thread(target=ak, args=(outer,), daemon=True)
    thr.start()
    fn, args, kw = ak.activity_request.get()
    out = fn(*args, **kw)
    assert val == out
    ak.activity_response.put(out)
    thr.join(0.1)  # Will raise if ak still waits


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
    message = comm.receive_activity()
    assert message['name'] == sub.__name__
    assert message['args'] == [str(v) for v in args]
    assert message['kw'] == {k: str(v) for k, v in kw.items()}
    fn, args, kw = message['data']
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
    assert message['payload']['names'] == names


    # Done
    comm.send_done()
    data = go.recv(2048)
    assert data, 'no data'
    message = json.loads(data)
    assert message['type'] == ak_runner.MessageType.done
