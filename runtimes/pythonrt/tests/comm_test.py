import json
from socket import socketpair

from ak_runner.comm import Comm, MessageType, format_traceback


def test_comm():
    go, py = socketpair()

    # Callback
    comm = Comm(py)
    fn_name, args, kw = "sub", (1, 7), {"verbose": False}
    comm.send_activity(fn_name, args, kw)
    data = go.recv(2048)
    assert data, "no data"

    go.sendall(data)
    payload = comm.recv(MessageType.callback)["payload"]
    assert payload["name"] == fn_name
    assert payload["args"] == [str(v) for v in args]
    assert payload["kw"] == {k: str(v) for k, v in kw.items()}

    # Module
    names = ["a", "c", "f"]
    comm.send_exported(names)
    data = go.recv(2048)
    assert data, "no data"
    message = json.loads(data)
    assert message["type"] == MessageType.module
    assert message["payload"]["entries"] == names

    # Done
    comm.send_done()
    data = go.recv(2048)
    assert data, "no data"
    message = json.loads(data)
    assert message["type"] == MessageType.done


def func_that_errs():
    json.loads(None)


def test_format_traceback():
    try:
        func_that_errs()
    except Exception as err:
        tb = format_traceback(err)

    assert len(tb) == 3
    assert "json" in tb[2]["file"]
