"""
Tests for the main module.

Need much more:
- Integration tests running actual workflow with temporal
- Bad input tests
"""

import builtins
import json
import os
import pickle
import sys
import traceback
from concurrent.futures import Future
from subprocess import Popen, TimeoutExpired, run
from threading import Event
from unittest.mock import MagicMock
from uuid import uuid4

import main
import pb.autokitteh.user_code.v1.runner_svc_pb2 as runner_pb
import pb.autokitteh.user_code.v1.user_code_pb2 as user_code
import pb.autokitteh.values.v1.values_pb2 as pb_values
import pytest
import values
from conftest import clear_module_cache, workflows
from mock_worker import MockWorker


def new_test_runner(code_dir, worker=None, server=None):
    runner = main.Runner(
        id="runner1",
        worker=worker,
        code_dir=code_dir,
        server=server,
    )
    runner._inactivity_timer.cancel()
    return runner


def test_help():
    cmd = [sys.executable, "main.py", "-h"]
    out = run(cmd)
    assert out.returncode == 0


def test_start(monkeypatch):
    mod_name = "program"
    clear_module_cache(mod_name)

    runner = new_test_runner(workflows.simple, worker=MagicMock())

    event_data = json.dumps(
        {
            "data": {"body": {"path": "/info", "method": "GET"}},
            "session_id": "ses_meow",
        }
    )
    event = user_code.Event(data=event_data.encode())
    entry_point = f"{mod_name}.py:on_event"
    req = runner_pb.StartRequest(entry_point=entry_point, event=event)
    context = MagicMock()

    # Restore print after this test
    monkeypatch.setattr(builtins, "print", print)

    resp = runner.Start(req, context)

    assert resp.error == ""
    assert not context.abort.called


def sub(a, b):
    return a - b


def test_execute():
    runner = new_test_runner(workflows.simple)

    class Worker:
        def __init__(self):
            self.event = Event()

        def ExecuteReply(self, msg):
            self.msg = msg
            self.event.set()
            return MagicMock()

    runner.worker = Worker()
    runner.activity_call = main.Call(sub, [1, 7], {}, Future())

    req = runner_pb.ExecuteRequest()
    resp = runner.Execute(req, None)
    assert resp.error == ""

    triggered = runner.worker.event.wait(1)
    assert triggered, "timeout waiting for worker event"

    result = pickle.loads(runner.worker.msg.result.custom.data)
    assert result.value == -6


def test_activity_reply():
    runner = new_test_runner(workflows.simple)
    fut = Future()
    runner.activity_call = main.Call(print, (), {}, fut)
    result = main.Result(42, None, None)
    req = runner_pb.ActivityReplyRequest(
        result=pb_values.Value(
            custom=pb_values.Custom(
                executor_id=runner.id,
                data=pickle.dumps(result),
                value=values.safe_wrap(result.value),
            ),
        )
    )
    resp = runner.ActivityReply(req, MagicMock())
    assert resp.error == ""
    assert fut.done()
    assert fut.result() == result.value


# TODO: This test takes about 14 seconds to finish, can we do it faster?
def test_start_timeout(tmp_path):
    timeout = 1
    cmd = [
        sys.executable,
        "main.py",
        "--skip-check-worker",
        "--port", "0",
        "--runner-id", "r1",
        "--code-dir", str(tmp_path),
        "--start-timeout", str(timeout),
    ]  # fmt: skip

    timeout = timeout + main.SERVER_GRACE_TIMEOUT + 1
    p = Popen(cmd)
    try:
        p.wait(timeout)
    except TimeoutExpired:
        p.kill()
        assert False, "server did not terminate"


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

    runner = new_test_runner(workflows.simple)
    text = runner.result_error(err)

    assert msg in text
    for name in ("fn_a", "fn_b", "fn_c"):
        assert name in text


class SlackError(Exception):
    def __init__(self, message, response):
        self.response = response
        super().__init__(message)


class AuthError(Exception):
    def __init__(self, message, *, response, body):
        self.response = response
        self.body = body
        super().__init__(message)


pickle_cases = [
    pytest.param(
        SlackError("cannot connect", response={"error": "bad token"}), id="args"
    ),
    pytest.param(
        AuthError("bad creds", response={"error": "no password"}, body=b"banana"),
        id="kw",
    ),
]


@pytest.mark.parametrize("err", pickle_cases)
def test_pickle_exception(err):
    def fn():
        raise err

    runner = new_test_runner("/tmp")
    result = runner._call(fn, [], {})
    data = pickle.dumps(result)
    result2 = pickle.loads(data)
    error = main.restore_error(result2.error)
    assert isinstance(error, err.__class__)


def test_activity():
    runner = new_test_runner(workflows.activity)
    worker = MockWorker(runner)
    event = json.dumps({"data": {"cat": "mitzi"}})
    worker.start("program.py:on_event", event.encode())


code = """
import json

def main():
    a()

def a():
    b()

def b():
    json.loads("{")  # Will raise
"""


def test_pb_traceback(monkeypatch, tmp_path):
    mod_name = f"program_{uuid4().hex}"
    py_file = tmp_path / f"{mod_name}.py"
    with open(py_file, "w") as out:
        out.write(code)

    monkeypatch.setattr(sys, "path", sys.path + [str(tmp_path)])

    mod = __import__(mod_name)
    try:
        mod.main()
    except ValueError as err:
        tb = main.TracebackException.from_exception(err)

    pb_tb = main.pb_traceback(main.tb_stack(tb))
    assert [f.filename for f in pb_tb] == [f.filename for f in tb.stack]


ftb_user_dir = "/tmp/user"
Frame = traceback.FrameSummary

ftb_cases = [
    # stack, index
    # runner then user
    (
        [
            Frame("/tmp/runner/main.py", 0, "user"),
            Frame(f"{ftb_user_dir}/program.py", 0, "user"),
        ]
        + sys.path,
        1,
    ),
    # user first
    (
        [
            Frame(f"{ftb_user_dir}/program.py", 0, "user"),
            Frame("/tmp/runner/main.py", 0, "user"),
        ],
        0,
    ),
    # no user
    (
        [
            Frame("/tmp/runner/main.py", 0, "user"),
            Frame("/tmp/runner/main.py", 0, "user"),
        ],
        0,
    ),
    # empty
    ([], 0),
]


@pytest.mark.parametrize("stack, index", ftb_cases)
def test_filter_tb(stack, index):
    out = main.filter_traceback(stack, ftb_user_dir)
    assert stack[index:] == out


def test_tb_stack():
    def a():
        b()

    def b():
        c()

    def c():
        raise ValueError("oops")

    try:
        a()
    except ValueError as e:
        err = e

    tb = main.TracebackException.from_exception(err)
    tb.stack[0].colno = lambda: 1  # unpickleable, not in Frame
    stack = main.tb_stack(tb)
    assert len(stack) == 4
    pickle.dumps(stack)

    pbt = main.pb_traceback(stack)
    assert len(pbt) == 4


def test_obj_callable():
    worker = MagicMock()
    worker.Activity.return_value = worker
    worker.error = None

    runner = new_test_runner(workflows.simple, worker=worker)

    class Adder:
        def __call__(self, a, b):
            return a + b

    fn = Adder()
    runner.start_activity(fn, (), {})


@pytest.mark.parametrize(
    "workflow",
    [
        pytest.param(workflows.async_activity, id="async_activity"),
        pytest.param(workflows.async_handler, id="async_handler"),
    ],
)
def test_async(workflow, capsys):
    clear_module_cache("program")
    runner = new_test_runner(workflow)
    worker = MockWorker(runner)
    runner.worker = worker

    event = json.dumps({"data": {"cat": "mitzi"}})
    worker.start("program.py:on_event", event.encode())
    worker.event.wait(1)

    assert worker.calls.get("ACTIVITY")

    captured = capsys.readouterr()
    assert "on_event: end (out=8)" in captured.out


def test_async_exc(monkeypatch):
    clear_module_cache("program")
    runner = new_test_runner(workflows.async_exc, server=MagicMock())
    worker = MockWorker(runner)
    runner.worker = worker

    def exit():
        raise SystemExit(1)

    monkeypatch.setattr(os, "_exit", exit)

    event = json.dumps({"data": {"cat": "mitzi"}})
    worker.start("program.py:on_event", event.encode())
    worker.event.wait(1)

    assert worker.calls.get("ACTIVITY")
