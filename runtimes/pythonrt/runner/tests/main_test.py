"""
Tests for the main module.

Need much more:
- Integration tests running actual workflow with temporal
- Bad input tests
"""

import json
import pickle
import sys
from concurrent.futures import Future
from subprocess import run
from unittest.mock import MagicMock

import main
import pb.autokitteh.user_code.v1.runner_svc_pb2 as runner_pb
import pb.autokitteh.user_code.v1.user_code_pb2 as user_code
import pb.autokitteh.values.v1.values_pb2 as pb_values
from conftest import clear_module_cache, workflows


def test_help():
    cmd = [sys.executable, "main.py", "-h"]
    out = run(cmd)
    assert out.returncode == 0


def test_start():
    mod_name = "program"
    clear_module_cache(mod_name)

    runner = main.Runner(
        id="runner1",
        worker=MagicMock(),
        code_dir=workflows.simple,
        server=None,
    )

    event_data = json.dumps({"body": {"path": "/info", "method": "GET"}})
    event = user_code.Event(data=event_data.encode())
    entry_point = f"{mod_name}.py:on_event"
    req = runner_pb.StartRequest(entry_point=entry_point, event=event)
    context = MagicMock()
    resp = runner.Start(req, context)

    assert resp.error == ""
    assert not context.abort.called


def sub(a, b):
    return a - b


def test_execute():
    runner = main.Runner(
        id="runner1",
        worker=None,
        code_dir=workflows.simple,
        server=None,
    )

    runner.activity_call = main.Call(sub, [1, 7], {}, Future())
    req = runner_pb.ExecuteRequest()
    resp = runner.Execute(req, None)
    assert resp.error == ""
    result = pickle.loads(resp.result.custom.data)
    assert result.value == -6


def test_activity_reply():
    runner = main.Runner(
        id="runner1",
        worker=None,
        code_dir=workflows.simple,
        server=None,
    )
    fut = Future()
    runner.activity_call = main.Call(print, (), {}, fut)
    result = main.Result(42, None, None)
    req = runner_pb.ActivityReplyRequest(
        result=pb_values.Value(
            custom=pb_values.Custom(
                executor_id=runner.id,
                data=pickle.dumps(result),
                value=main.safe_wrap(result.value),
            ),
        )
    )
    resp = runner.ActivityReply(req, MagicMock())
    assert resp.error == ""
    assert fut.done()
    assert fut.result() == result.value


def test_result_error():
    msg = "oops"

    def fn_a():
        fn_b()

    def fn_b():
        fn_c()

    def fn_c():
        raise ZeroDivisionError(msg)

    err = None
    try:
        fn_a()
    except ZeroDivisionError as e:
        err = e

    text = main.result_error(err)

    assert msg in text
    for name in ("fn_a", "fn_b", "fn_c"):
        assert name in text


class SlackError(Exception):
    def __init__(self, message, response):
        self.response = response
        super().__init__(message)


def test_pickle_exception():
    def fn():
        raise SlackError("cannot connect", response={"error": "bad token"})

    runner = main.Runner("r1", None, "/tmp", None)
    result = runner._call(fn, [], {})
    data = pickle.dumps(result)
    result2 = pickle.loads(data)
    assert isinstance(result2.error, SlackError)
