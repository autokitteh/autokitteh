from unittest.mock import MagicMock

import syscalls


def test_ak_unsubscribe(monkeypatch):
    mock = MagicMock()
    monkeypatch.setattr(syscalls, "call_grpc", mock)

    rid = "r1"
    sc = syscalls.SysCalls(rid, mock, mock)
    sid = "s1"
    sc.ak_unsubscribe(sid)
    assert mock.called
    req = mock.call_args[0][2]
    assert req.runner_id == rid
    assert req.signal_id == sid
