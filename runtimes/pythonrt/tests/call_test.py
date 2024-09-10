import json
import pickle
import re
from base64 import b64encode
from socket import socketpair
from threading import Thread
from time import sleep
from types import ModuleType
from unittest.mock import MagicMock

import pytest
from autokitteh import activity, decorators
from conftest import testdata
from loader_test import simple_dir

import ak_runner


# Used by test_nested, must be global to be pickled.
def ak(fn):
    pass


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
            "type": ak_runner.MessageType.callback,
        },
        {
            "type": ak_runner.MessageType.response,
            "payload": {
                "value": pickle.dumps(val),
            },
        },
    ]
    comm.extract_activity.side_effect = [
        {
            "name": "outer",
            "args": [],
            "kw": {},
            "data": (outer, (), {}),
        },
    ]

    akc = ak_runner.AKCall(comm)
    akc.set_module(json)  # outer & inner should look external
    akc(outer)

    comm.send_activity.assert_called_once()


def in_act_2(v):
    print(f"in_act_2: {v}")


def in_act_1(v):
    print("in_act_1: in")
    in_act_2(v)
    print("in_act_2: in")


def test_in_activity():
    class Comm:
        def __init__(self):
            self.values = []
            self.num_activities = 0
            self.n = 0

        def send_activity(self, func, args, kw):
            self.num_activities += 1
            self.message = {"data": (func, args, kw)}

        def send_response(self, value):
            self.values.append(value)

        def extract_response(self, msg):
            return msg["payload"]["value"]

        def recv(self, *types):
            self.n += 1

            if self.n == 1:
                return {
                    "type": ak_runner.MessageType.callback,
                    "payload": self.message,
                }

            return {
                "type": ak_runner.MessageType.response,
                "payload": {"value": pickle.dumps(self.values[0])},
            }

        def extract_activity(self, msg):
            return msg["payload"]

    comm = Comm()
    akc = ak_runner.AKCall(comm)
    akc.set_module(json)  # in_act_1 should look external
    akc(in_act_1, 7)
    assert comm.num_activities == 1

    akc(in_act_1, 6)
    assert comm.num_activities == 2


sleep_code = """
from time import sleep
import time

def handler(event):
    print('before')
    sleep(1)
    time.sleep(2)
    print('after')
"""


def test_sleep(tmp_path):
    mod_name = "sleeper"

    file_name = tmp_path / (mod_name + ".py")
    with open(file_name, "w") as out:
        out.write(sleep_code)

    comm = MagicMock()

    ak_call = ak_runner.AKCall(comm)
    mod = ak_runner.load_code(tmp_path, ak_call, mod_name)
    ak_call.set_module(mod)
    event = {"type": "login", "user": "puss"}
    mod.handler(event)
    assert comm.send_call.call_count == 2


def test_activity():
    mod_name = "activity"
    mod = ak_runner.load_code(testdata, lambda f: f, mod_name)
    fn = mod.phone_home
    assert getattr(fn, decorators.ACTIVITY_ATTR, False)


def mock_tp_go(sock):
    """Mock Go server for test_pickle_function"""
    fp = sock.makefile("r")
    fp.readline()

    # Mock replay
    result = b64encode(pickle.dumps(None))
    message = {
        "type": ak_runner.MessageType.response,
        "payload": {"value": result.decode()},
    }
    out = json.dumps(message) + "\n"
    try:
        sock.sendall(out.encode())
    except BrokenPipeError:
        pass


def test_pickle_function():
    go, py = socketpair()
    Thread(target=mock_tp_go, args=(go,), daemon=True).start()

    root_path = str(simple_dir)
    comm = ak_runner.Comm(py)
    ak_call = ak_runner.AKCall(comm)
    mod = ak_runner.load_code(root_path, ak_call, "simple")
    ak_call.module = mod
    event = {
        "data": {
            "body": b'{"name": "grumpy", "type": "cat"}',
        },
    }

    ak_call(mod.printer, event)


def test_sleep_activity():
    comm = MagicMock()
    ak_call = ak_runner.AKCall(comm)
    ak_call.in_activity = True
    ak_call(sleep, 0.1)

    assert comm.send_call.call_count == 0


def test_call_non_func():
    comm = MagicMock()
    ak_call = ak_runner.AKCall(comm)
    with pytest.raises(ValueError):
        ak_call("hello")


class List(list):
    pass


def test_should_run_as_activity():
    mod_name = "ak_test_module_name"
    mod = ModuleType(mod_name)

    ak_call = ak_runner.AKCall(None)

    def fn():
        pass

    # loading
    assert not ak_call.should_run_as_activity(fn)
    ak_call.set_module(mod)

    # after loading
    assert ak_call.should_run_as_activity(fn)

    # Function from same module
    fn.__module__ = mod_name
    assert not ak_call.should_run_as_activity(fn)

    # Marked activity
    fn = activity(fn)
    assert ak_call.should_run_as_activity(fn)

    # In activity
    ak_call.in_activity = True
    assert not ak_call.should_run_as_activity(fn)
    ak_call.in_activity = False

    # Deterministic
    assert not ak_call.should_run_as_activity(re.compile)

    # Subclass of built-in type
    lst = List()
    assert not ak_call.should_run_as_activity(lst.count)
