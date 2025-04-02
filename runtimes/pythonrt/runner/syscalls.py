"""AutoKitteh syscalls

Note: SysCalls.ak_* methods (e.g. ak_start) signature must match the signature of the
matching function in autokitteh (e.g. autokitteh.start).
"""

import json
import os
from datetime import timedelta

import grpc
import log
import pb.autokitteh.user_code.v1.handler_svc_pb2 as pb
from autokitteh import AttrDict, Signal
from autokitteh.decorators import ACTIVITY_ATTR
import values


class SyscallError(Exception):
    pass


def mark_no_activity(fn):
    """Mark that a function should not run as activity."""
    setattr(fn, ACTIVITY_ATTR, False)
    return fn


def _timeout_arg_into_ms(timeout: timedelta | int | float) -> int:
    if not timeout:
        return 0

    if isinstance(timeout, int | float):
        return int(timeout * 1000)
    elif isinstance(timeout, timedelta):
        return int(timeout.total_seconds() * 1000)

    raise TypeError(f"timeout {timeout!r} should be a timedelta or number of seconds")


class SysCalls:
    def __init__(self, runner_id, worker, log):
        self.runner_id = runner_id
        self.worker = worker
        self.log = log
        self.mark_ak_no_activity()

    def ak_start(
        self,
        loc: str,
        data: dict | None = None,
        memo: dict | None = None,
        project: str | None = None,
    ) -> str:
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
            project=project or "",
        )
        resp = call_grpc("start", self.worker.StartSession, req)
        return resp.session_id

    def ak_sleep(self, seconds):
        log.info("ak_sleep: %r", seconds)
        if seconds < 0:
            raise ValueError("negative secs")

        req = pb.SleepRequest(
            runner_id=self.runner_id,
            duration_ms=int(seconds * 1000),
        )

        call_grpc("sleep", self.worker.Sleep, req)

    def ak_subscribe(self, source: str, filter: str = "") -> str:
        log.info("ak_subscribe: %r %r", source, filter)
        if not source:
            raise ValueError("missing source")

        req = pb.SubscribeRequest(
            runner_id=self.runner_id, connection=source, filter=filter
        )

        resp = call_grpc("subscribe", self.worker.Subscribe, req)
        return resp.signal_id

    def ak_next_event(self, subscription_id, *, timeout=None):
        log.info("ak_next_event: %r %r", subscription_id, timeout)

        ids = subscription_id
        if isinstance(ids, str):
            ids = [ids]

        req = pb.NextEventRequest(
            runner_id=self.runner_id,
            signal_ids=ids,
            timeout_ms=_timeout_arg_into_ms(timeout),
        )

        resp = call_grpc("next_event", self.worker.NextEvent, req)

        try:
            data = json.loads(resp.event.data)
        except (ValueError, TypeError, AttributeError) as err:
            raise SyscallError(f"next_event: invalid event: {err}")

        return AttrDict(data) if isinstance(data, dict) else data

    def ak_unsubscribe(self, subscription_id):
        log.info("ak_unsubscribe: %r", subscription_id)
        req = pb.UnsubscribeRequest(runner_id=self.runner_id, signal_id=subscription_id)
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

    def ak_signal(self, session_id: str, name: str, payload: any = None) -> None:
        log.info("signal: %r %r", session_id, name)

        req = pb.SignalRequest(
            runner_id=self.runner_id,
            session_id=session_id,
            signal=pb.Signal(
                name=name,
                payload=values.wrap(payload),
            ),
        )

        call_grpc("signal", self.worker.Signal, req)

    def ak_next_signal(
        self, name: str | list[str], *, timeout: timedelta | int | float = None
    ) -> Signal:
        log.info("ak_next_signal: %r %r", name, timeout)

        names = name
        if isinstance(names, str):
            names = [names]

        req = pb.NextSignalRequest(
            runner_id=self.runner_id,
            names=names,
            timeout_ms=_timeout_arg_into_ms(timeout),
        )

        resp = call_grpc("next_signal", self.worker.NextSignal, req)

        sig = resp.signal

        if sig and sig.name:
            return Signal(
                name=sig.name,
                payload=values.unwrap(sig.payload),
            )

        return None

    @classmethod
    def mark_ak_no_activity(cls):
        """Mark ak_* methods as not activity."""
        for attr, value in cls.__dict__.items():
            if not attr.startswith("ak_") or not callable(value):
                continue

            mark_no_activity(value)


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
