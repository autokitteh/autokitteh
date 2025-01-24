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
from subprocess import Popen, TimeoutExpired, run
from unittest.mock import MagicMock

import main
import pb.autokitteh.user_code.v1.runner_svc_pb2 as runner_pb
import pb.autokitteh.user_code.v1.user_code_pb2 as user_code
import pb.autokitteh.values.v1.values_pb2 as pb_values
import values
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


def new_test_runner(code_dir):
    runner = main.Runner(
        id="runner1",
        worker=None,
        code_dir=code_dir,
        server=None,
    )
    runner._inactivty_timer.cancel()
    return runner


def test_execute():
    runner = new_test_runner(workflows.simple)
    runner.activity_call = main.Call(sub, [1, 7], {}, Future())
    req = runner_pb.ExecuteRequest()
    resp = runner.Execute(req, None)
    assert resp.error == ""
    result = pickle.loads(resp.result.custom.data)
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
    cmd = [
        sys.executable,
        "main.py",
        "--skip-check-worker",
        "--port",
        "0",
        "--runner-id",
        "r1",
        "--code-dir",
        str(tmp_path),
    ]

    timeout = main.START_TIMEOUT + main.SERVER_GRACE_TIMEOUT + 1
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

    runner = new_test_runner("/tmp")
    result = runner._call(fn, [], {})
    data = pickle.dumps(result)
    result2 = pickle.loads(data)
    assert isinstance(result2.error, SlackError)
