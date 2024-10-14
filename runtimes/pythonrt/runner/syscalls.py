"""AutoKitteh syscalls

Convert general func(*args, **kw) to a specific gRPC call to worker.
"""

import json
import os

import grpc

from autokitteh import AttrDict
import log
import pb.autokitteh.remote.v1.remote_pb2 as pb
import pb.autokitteh.remote.v1.remote_pb2_grpc as rpc


class SyscallError(Exception):
    pass


class SysCalls:
    def __init__(self, runner_id, worker):
        self.runner_id = runner_id
        self.worker: rpc.WorkerStub = worker

        self.ak_funcs = {
            "next_event": self.ak_next_event,
            "sleep": self.ak_sleep,
            "subscribe": self.ak_subscribe,
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

    def ak_next_event(self, args, kw):
        (id,) = extract_args(["subscription_id"], args, kw)
        if not id:
            raise ValueError("empty subscription_id")
        req = pb.NextEventRequest(runner_id=self.runner_id, signal_ids=[id])

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

    def ak_refresh_oauth(integration: str, self, connection: str, scopes: list[str]):
        # TODO BEFORE MERGING: REMOVE PRINTS
        print("!!!!!!!!!! refresh_oauth IS overriden !!!!!!!!!!")
        print("@@@@@@@@@@ connection: ", connection)
        print("########## integration: ", integration)
        print("$$$$$$$$$$ scopes: ", scopes)
        req = pb.RefreshRequest(
            runner_id=self.runner_id,
            connection=connection,
            integration=integration,
            scopes=scopes,
        )
        resp: pb.RefreshResponse = call_grpc(
            "refresh_oauth", self.worker.RefreshOAuthToken, req
        )
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
    """
    values = []
    for i, name in enumerate(names):
        v = args[i] if i < len(args) else kw.get(name, _missing)
        if v is _missing:
            raise ValueError(f"missing {name!r}")
        values.append(v)

    return values
