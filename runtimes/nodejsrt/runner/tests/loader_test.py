import ast
import json
from textwrap import dedent
from unittest.mock import MagicMock

import call
import loader
import pytest
from autokitteh import AttrDict
from conftest import clear_module_cache, workflows


def test_load_code():
    mod_name = "program"
    clear_module_cache(mod_name)

    calls = []

    class MockCall(call.AKCall):
        def __init__(self, *args, **kw):
            super().__init__(*args, **kw)

        def __call__(self, fn, *args, **kw):
            if not self.should_run_as_activity(fn):
                return fn(*args, **kw)

            if fn.__module__ != mod_name:
                calls.append((fn, args, kw))
            return fn(*args, **kw)

    code_dir = workflows.simple
    ak_call = MockCall(runner=MagicMock(), code_dir=code_dir)

    mod = loader.load_code(code_dir, ak_call, mod_name)
    ak_call.set_module(mod)
    fn = getattr(mod, "on_event", None)
    assert fn, "on_event not found"

    event = AttrDict({"user": "elliot", "action": "login"})
    fn(event)  # Make sure it runs without error.


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


def test_exports():
    code_dir, file_name = workflows.simple, "program.py"
    exports = list(loader.exports(code_dir, file_name))
    expected = {
        "file": file_name,
        "line": 6,
        "name": "on_event",
        "args": ["event"],
    }
    assert exports == [expected]


def test_multi_file():
    mod_name = "handler"
    clear_module_cache(mod_name, "hlog")

    runner = MagicMock()

    code_dir = workflows.multi_file
    ak_call = call.AKCall(runner=runner, code_dir=code_dir)

    mod = loader.load_code(code_dir, ak_call, mod_name)
    ak_call.set_module(mod)
    fn = getattr(mod, "on_event", None)
    assert fn, "on_event not found"

    event = AttrDict({"data": json.dumps({"user": "joe", "action": "login"})})
    fn(event)  # Make sure it runs without error.
    assert runner.call_in_activity.call_count == 2


def test_class_args():
    code = """
    class Player:
        def move(self, dx, dy):
            self.x += dx
            self.y += dy

        def __init__(self, x, y):
            self.x, self.y = x, y
    """
    tree = ast.parse(dedent(code))
    node = tree.body[0]
    args = loader.class_args(node)
    assert args == ["x", "y"]


fn_args_cases = [
    pytest.param(
        """
        def fn():
            pass
        """,
        [],
        id="no args",
    ),
    pytest.param(
        """
        def inc(n):
            return n + 1
        """,
        ["n"],
        id="single argument",
    ),
    pytest.param(
        """
        def add(a, b, **kw):
            return a + b
        """,
        ["a", "b", "kw"],
        id="kw",
    ),
]


@pytest.mark.parametrize("code, args", fn_args_cases)
def test_fn_args(code, args):
    code = dedent(code)
    tree = ast.parse(code)
    out = loader.fn_args(tree.body[0])
    assert out == args
