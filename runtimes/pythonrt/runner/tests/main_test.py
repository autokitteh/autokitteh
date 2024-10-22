"""
Tests for the main module.

Need much more:
- Integration tests running actual workflow with temporal
- Bad input tests
"""

from concurrent.futures import Future
from subprocess import run
from sys import executable
from uuid import uuid4

from conftest import workflows

import pb.autokitteh.remote.v1.remote_pb2 as pbremote
import pb.autokitteh.values.v1.values_pb2 as pbvalues
from values import Wrapper, wrap
import main

_XID = "runner1"


def test_help():
    cmd = [executable, "main.py", "-h"]
    out = run(cmd)
    assert out.returncode == 0


def test_start():
    runner = main.Runner(
        id=_XID,
        worker=None,
        code_dir=workflows.simple,
        server=None,
    )

    event_data = {"body": wrap({"path": "/info", "method": "GET"})}
    event = pbremote.Event(data=event_data)
    req = pbremote.StartRequest(entry_point="program.py:on_event", event=event)
    resp = runner.Start(req, None)
    assert resp.error == ""


def sub(a, b):
    return a - b


def test_execute():
    runner = main.Runner(
        id=_XID,
        worker=None,
        code_dir=workflows.simple,
        server=None,
    )

    call_id = uuid4().hex
    runner.calls[call_id] = (sub, [1, 7], {})
    req = pbremote.ExecuteRequest(
        value=pbvalues.Value(
            function=pbvalues.Function(
                data=call_id.encode(),
            )
        ),
    )
    resp = runner.Execute(req, None)
    assert resp.error == ""

    v = Wrapper(_XID).unwrap(resp.result)
    assert v == -6


def test_activity_reply():
    runner = main.Runner(
        id=_XID,
        worker=None,
        code_dir=workflows.simple,
        server=None,
    )
    call_id = uuid4().hex
    fut = Future()
    runner.replies[call_id] = fut
    value = 42
    req = pbremote.ActivityReplyRequest(
        data=call_id.encode(),
        result=wrap(value),
    )

    resp = runner.ActivityReply(req, None)
    assert resp.error == ""
    assert fut.done()
    assert fut.result() == value
