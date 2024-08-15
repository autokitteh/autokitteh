import re
from time import sleep
from types import ModuleType
from unittest.mock import MagicMock

import pytest
from autokitteh import activity
from conftest import workflows

import call
import loader


def test_sleep():
    runner = MagicMock()
    ak_call = call.AKCall(runner)
    mod_name = "program"

    mod = loader.load_code(workflows.sleeper, ak_call, mod_name)
    ak_call.set_module(mod)
    event = {"type": "login", "user": "puss"}
    mod.handler(event)
    assert runner.syscall.call_count == 2


def test_sleep_activity():
    comm = MagicMock()
    ak_call = call.AKCall(comm)
    ak_call.in_activity = True
    ak_call(sleep, 0.1)

    assert comm.send_call.call_count == 0


def test_call_non_func():
    comm = MagicMock()
    ak_call = call.AKCall(comm)
    with pytest.raises(ValueError):
        ak_call("hello")


def test_should_run_as_activity():
    mod_name = "ak_test_module_name"
    mod = ModuleType(mod_name)

    ak_call = call.AKCall(None)

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
