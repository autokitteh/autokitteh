import ast
import json
from unittest.mock import MagicMock

import pytest
from autokitteh import AttrDict
from conftest import workflows, clear_module_cache

import call
import loader


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
    assert exports == ["on_event"]


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
