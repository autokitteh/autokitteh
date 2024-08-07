import ast
from os import environ
from socket import socket
from unittest.mock import MagicMock

import pytest
from conftest import testdata

import ak_runner
from ak_runner import loader

simple_dir = testdata / "simple"


def test_load_code():
    calls = []

    mod_name = "mod"

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

    mod = ak_runner.load_code(testdata, ak_call, mod_name)
    ak_call.set_module(mod)
    fn = getattr(mod, "parse", None)
    assert fn, "parse not found"

    out = fn("meow")
    assert out == 7
    assert len(calls) == 1, "calls"
    fn = calls[0][0]
    assert fn.__qualname__ == "datetime.now"


def test_load_twice(tmp_path):
    mod_name = "x"
    file_name = tmp_path / (mod_name + ".py")
    var, val = "x", 1
    with open(file_name, "w") as out:
        print(f"{var} = {val}", file=out)

    mod = ak_runner.load_code(tmp_path, lambda x: x, mod_name)
    assert getattr(mod, var) == val

    # See that module is not cached.
    val += 1
    with open(file_name, "w") as out:
        print(f"{var} = {val}", file=out)

    mod = ak_runner.load_code(tmp_path, lambda x: x, mod_name)
    assert getattr(mod, var) == val


def test_load_simple():
    root_path = str(simple_dir)

    def action_fn(*args, **kw):
        pass

    ak_runner.load_code(root_path, action_fn, "simple")


transform_cases = [
    # code, transformed
    ("get(1)", "_ak_call(get, 1)"),
    ('requests.get("https://go.dev")', '_ak_call(requests.get, "https://go.dev")'),
    (
        'sheets.values().get("A1:B4").execute()',
        '_ak_call(_ak_call(_ak_call(google.sheets.values).get, "A1:A10").execute)',
    ),
    ("add(get(1), get(2))", "_ak_call(add, _ak_call(get, 1), _ak_call(get, 2))"),
]


@pytest.mark.parametrize("code, transformed", transform_cases)
def test_transform(code, transformed):
    mod = ast.parse(code)
    trans = loader.Transformer("<stdin>", code)
    out = trans.visit(mod)
    assert transformed, ast.unparse(out)


def test_module_level():
    comm = MagicMock()
    akc = ak_runner.AKCall(comm)
    mod = ak_runner.load_code(testdata, akc, "modlevel")
    assert mod.home == environ["HOME"]
    akc.set_module(mod)

    mod.on_event(None)
    assert mod.ncalls == 1


const_code = """
def handler(event):
    msg = 'INFO: event: {event!r}'.format(event=event)
    print(msg)
    return msg
"""


def test_const(tmp_path):
    mod_name = "const"
    file_name = tmp_path / (mod_name + ".py")
    with open(file_name, "w") as out:
        out.write(const_code)

    def ak_call(fn, *args, **kw):
        return fn(*args, **kw)

    mod = loader.load_code(tmp_path, ak_call, mod_name)
    assert mod.handler("meow") == "INFO: event: 'meow'"
