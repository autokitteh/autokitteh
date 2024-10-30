"""AutoKitteh syscalls

Convert general func(*args, **kw) to a specific gRPC call to worker.
"""

import json
import os
from datetime import timedelta

import grpc
import log
import pb.autokitteh.remote.v1.remote_pb2 as pb
import pb.autokitteh.remote.v1.remote_pb2_grpc as rpc
from autokitteh import AttrDict


class SyscallError(Exception):
    pass


class SysCalls:
    def __init__(self, runner_id, worker, log):
        self.runner_id = runner_id
        self.worker: rpc.WorkerStub = worker
        self.log = log

        self.ak_funcs = {
            "next_event": self.ak_next_event,
            "sleep": self.ak_sleep,
            "start": self.ak_start,
            "subscribe": self.ak_subscribe,
            "unsubscribe": self.ak_unsubscribe,
        }

    def call(self, fn, args, kw):
        method = self.ak_funcs.get(fn.__name__)
        if method is None:
            raise ValueError(f"unknown ak function: {fn.__name__!r}")
        return method(args, kw)

    def ak_start(self, loc: str, data: dict = None, memo: dict = None) -> str:
        self.log.info("ak_start: %r", loc)
        data = {} if data is None else data
        memo = {} if memo is None else memo

        try:
            json_data = json.dumps(data)
        except (ValueError, TypeError) as err:
            raise ValueError(f"start: invalid data: {err}")

        try:
            json_memo = json.dumps(memo)
        except (ValueError, TypeError) as err:
            raise ValueError(f"start: invalid memo: {err}")

        req = pb.StartSessionRequest(
            runner_id=self.runner_id,
            loc=loc,
            data=json_data.encode(),
            memo=json_memo.encode(),
        )
        resp = call_grpc("start", self.worker.StartSession, req)
        return resp.session_id

    def ak_sleep(self, seconds):
        if seconds < 0:
            raise ValueError("negative secs")

        req = pb.SleepRequest(
            runner_id=self.runner_id,
            duration_ms=int(seconds * 1000),
        )

        call_grpc("sleep", self.worker.Sleep, req)

    def ak_subscribe(self, args, kw):
        log.info("ak_subscribe: %r %r", args, kw)
        connection_id, filter = extract_args(["connection_name", "filter"], args, kw)
        if not connection_id or not filter:
            raise ValueError("missing connection_id or filter")

        req = pb.SubscribeRequest(
            runner_id=self.runner_id, connection=connection_id, filter=filter
        )
        resp = call_grpc("subscribe", self.worker.Subscribe, req)
        return resp.signal_id

    def ak_next_event(self, subscription_id, *, timeout=None):
        log.info("ak_next_event: %r %r", subscription_id, timeout)

        ids = subscription_id
        if isinstance(ids, str):
            ids = [ids]

        req = pb.NextEventRequest(runner_id=self.runner_id, signal_ids=ids)
        if timeout:
            if isinstance(timeout, float) or isinstance(timeout, int):
                req.timeout_ms = int(timeout * 1000)
            elif isinstance(timeout, timedelta):
                req.timeout_ms = int(timeout.total_seconds() * 1000)
            else:
                raise TypeError(f"timeout should be timedelta or int, got {timeout!r}")

        resp = call_grpc("next_event", self.worker.NextEvent, req)

        try:
            data = json.loads(resp.event.data)
        except (ValueError, TypeError, AttributeError) as err:
            raise SyscallError(f"next_event: invalid event: {err}")

        return AttrDict(data) if isinstance(data, dict) else data

    def ak_unsubscribe(self, args, kw):
        (id,) = extract_args(["subscription_id"], args, kw)
        if not id:
            raise ValueError("empty subscription_id")

        req = pb.UnsubscribeRequest(runner_id=self.runner_id, signal_id=id)
        call_grpc("unsubscribe", self.worker.Unsubscribe, req)

    def ak_encode_jwt(self, payload: dict[str, int], connection: str, algorithm: str):
        req = pb.EncodeJWTRequest(
            runner_id=self.runner_id,
            payload=payload,
            connection=connection,
            algorithm=algorithm,
        )
        resp: pb.EncodeJWTResponse = call_grpc("encode_jwt", self.worker.EncodeJWT, req)
        if resp.error:
            raise SyscallError(f"encode_jwt: {resp.error}")
        return resp.jwt

    def ak_refresh_oauth(self, integration: str, connection: str):
        req = pb.RefreshRequest(
            runner_id=self.runner_id,
            integration=integration,
            connection=connection,
        )
        resp: pb.RefreshResponse = call_grpc(
            "refresh_oauth", self.worker.RefreshOAuthToken, req
        )
        if resp.error:
            raise SyscallError(f"refresh_oauth: {resp.error}")
        return resp.token, resp.expires.ToDatetime()


# Can't use None since it's a valid value
_missing = object()


def call_grpc(name, fn, args):
    try:
        resp = fn(args)
        if resp.error:
            raise SyscallError(f"{name}: {resp.error}")
        return resp
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAVAILABLE or grpc.StatusCode.CANCELLED:
            os._exit(1)
        raise e


def extract_args(names, args, kw):
    """Extract arguments from args and kw, will raise ValueError if missing.

    >>> extract_args(["id", "timeout"], ["sig1", 1.2], {})
    ['sig1', 1.2]
    >>> extract_args(["id", "timeout"], ["sig1"], {"timeout": 1.2})
    ['sig1', 1.2]
    >>> extract_args(["id", "timeout"], [], {"id": "sig1", "timeout": 1.2})
    ['sig1', 1.2]
    >>> extract_args(["id", "timeout?"], [], {"id": "sig1", "timeout": 1.2})
    ['sig1', 1.2]
    >>> extract_args(["id", "timeout?"], [], {"id": "sig1"})
    ['sig1', None]
    >>> extract_args(["id", "timeout?"], ["sig1"], {})
    ['sig1', None]
    """
    values = []
    for i, name in enumerate(names):
        optional = name.endswith("?")
        if optional:
            name = name[:-1]
        v = args[i] if i < len(args) else kw.get(name, _missing)
        if v is _missing:
            if not optional:
                raise ValueError(f"missing {name!r}")
            v = None
        values.append(v)

    return values
