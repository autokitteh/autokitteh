"""AutoKitteh syscalls

Convert general func(*args, **kw) to a specific gRPC call to worker.
"""

import json
from datetime import timedelta

import log
import remote_pb2 as pb
import remote_pb2_grpc as rpc
from autokitteh import AttrDict


class SyscallError(Exception):
    pass


class SysCalls:
    def __init__(self, runner_id, worker):
        self.runner_id = runner_id
        self.worker: rpc.WorkerStub = worker

        self.ak_funcs = {
            "sleep": self.ak_sleep,
            "subscribe": self.ak_subscribe,
            "next_event": self.ak_next_event,
            "unsubscribe": self.ak_unsubscribe,
        }

    def call(self, fn, args, kw):
        method = self.ak_funcs.get(fn.__name__)
        if method is None:
            raise ValueError(f"unknown ak function: {fn.__name__!r}")
        return method(args, kw)

    def ak_sleep(self, args, kw):
        log.info("ak_sleep: %r %r", args, kw)
        (secs,) = extract_args(["secs"], args, kw)
        if secs < 0:
            raise ValueError("negative secs")

        req = pb.SleepRequest(
            runner_id=self.runner_id,
            duration_ms=int(secs * 1000),
        )
        resp = self.worker.Sleep(req)
        if resp.error:
            raise SyscallError(f"sleep: {resp.error}")

    def ak_subscribe(self, args, kw):
        log.info("ak_subscribe: %r %r", args, kw)
        connection_id, filter = extract_args(["connection_name", "filter"], args, kw)
        if not connection_id or not filter:
            raise ValueError("missing connection_id or filter")

        req = pb.SubscribeRequest(
            runner_id=self.runner_id, connection=connection_id, filter=filter
        )
        resp = self.worker.Subscribe(req)
        if resp.error:
            raise SyscallError(f"subscribe: {resp.error}")
        return resp.signal_id

    def ak_next_event(self, args, kw):
        (id,) = extract_args(["subscription_id"], args, kw)
        if not id:
            raise ValueError("empty subscription_id")

        timeout = kw.get("timeout")
        if len(args) == 2:
            timeout = args[1]

        if timeout:
            if not isinstance(timeout, timedelta):
                raise TypeError(f"timeout should be timedelta, got {type(timeout)}")
            if timeout <= timedelta(0):
                raise ValueError(f"bad timeout: {timeout!r}")

        req = pb.NextEventRequest(runner_id=self.runner_id, signal_ids=[id])
        if timeout:
            req.timeout_ms = int(timeout.total_seconds() * 1000)
        resp = self.worker.NextEvent(req)
        if resp.error:
            raise SyscallError(f"next_event: {resp.error}")

        try:
            value = json.loads(resp.event.data)
        except (ValueError, TypeError, AttributeError) as err:
            raise SyscallError(f"next_event: invalid event: {err}")

        value = {} if value is None else value  # None means timeout
        if not isinstance(value, dict):
            raise TypeError(f"next_event returned {value!r}, expected dict")
        value = AttrDict(value)
        return value

    def ak_unsubscribe(self, *args, **kw):
        (id,) = extract_args(["subscription_id"], args, kw)
        if not id:
            raise ValueError("empty subscription_id")

        req = pb.UnsubscribeRequest(runner_id=self.runner_id, signal_id=id)
        resp = self.worker.Unsubscribe(req)
        if resp.error:
            raise SyscallError(f"unsubscribe: {resp.error}")


# Can't use None since it's a valid value
_missing = object()


def extract_args(names, args, kw):
    """Extract arguments from args and kw, will raise ValueError if missing.

    >>> extract_args(["id", "timeout"], ["sig1", 1.2], {})
    ['sig1', 1.2]
    >>> extract_args(["id", "timeout"], ["sig1"], {"timeout": 1.2})
    ['sig1', 1.2]
    >>> extract_args(["id", "timeout"], [], {"id": "sig1", "timeout": 1.2})
    ['sig1', 1.2]
    """
    values = []
    for i, name in enumerate(names):
        v = args[i] if i < len(args) else kw.get(name, _missing)
        if v is _missing:
            raise ValueError(f"missing {name!r}")
        values.append(v)

    return values
