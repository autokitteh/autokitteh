"""
Tests for the main module.

Need much more:
- Integration tests running actual workflow with temporal
- Bad input tests
"""

import json
import pickle
from concurrent.futures import Future
from subprocess import run
from sys import executable
from uuid import uuid4

from conftest import workflows

import remote_pb2 as pb
import main


def test_help():
    cmd = [executable, "main.py", "-h"]
    out = run(cmd)
    assert out.returncode == 0


def test_start():
    runner = main.Runner(
        id="runner1",
        worker=None,
        code_dir=workflows.simple,
    )

    event_data = json.dumps({"body": {"path": "/info", "method": "GET"}})
    event = pb.Event(data=event_data.encode())
    req = pb.StartRequest(run_id="run1", entry_point="program.py:on_event", event=event)
    resp = runner.Start(req, None)
    assert resp.error == ""


def sub(a, b):
    return a - b


def test_execute():
    runner = main.Runner(
        id="runner1",
        worker=None,
        code_dir=workflows.simple,
    )

    call_id = uuid4().hex
    runner.calls[call_id] = (sub, [1, 7], {})
    req = pb.ExecuteRequest(call_id=call_id)
    resp = runner.Execute(req, None)
    assert resp.error == ""
    value = pickle.loads(resp.result)
    assert value == -6


def test_activity_reply():
    runner = main.Runner(
        id="runner1",
        worker=None,
        code_dir=workflows.simple,
    )
    call_id = uuid4().hex
    fut = Future()
    runner.replies[call_id] = fut
    value = 42
    req = pb.ActivityReplyRequest(
        call_id=call_id,
        result=pickle.dumps(value, protocol=0),
    )
    resp = runner.ActivityReply(req, None)
    assert resp.error == ""
    assert fut.done()
    assert fut.result() == value
