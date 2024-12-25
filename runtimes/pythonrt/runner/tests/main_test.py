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
