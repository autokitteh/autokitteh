"""A mock worker so we can debug the runner without ak."""

import asyncio
import sys
from itertools import count
from unittest.mock import MagicMock
from collections import Counter

import pb

_next_id = count(1).__next__


def next_signal_id():
    return f"signal-{_next_id()}"


class MockWorker(pb.handler_rpc.HandlerService):
    def __init__(self, runner: pb.runner_rpc.RunnerService, verbose=False):
        self.runner = runner
        self.runner.worker = self
        self.event = asyncio.Event()
        self.calls = Counter()
        self.verbose = verbose

    async def start(self, entry_point, data):
        req = pb.runner.StartRequest(
            entry_point=entry_point, event=pb.user_code.Event(data=data)
        )
        ctx = MagicMock()
        resp: pb.runner.StartResponse = await self.runner.Start(req, ctx)
        self.log("START RESPONSE:", resp)
        await self.event.wait()

    async def call_execute(self, data):
        ctx = MagicMock()

        await asyncio.sleep(0.1)
        req = pb.runner.ExecuteRequest(data=data)
        resp: pb.runner.ExecuteResponse = await self.runner.Execute(req, ctx)
        self.log("EXECUTE RESPONSE:", resp)

    async def call_activity_reply(self, msg: pb.handler.ExecuteReplyRequest):
        ctx = MagicMock()

        req = pb.runner.ActivityReplyRequest(
            result=msg.result,
            error=msg.error,
            # TODO: Traceback?
        )
        resp: pb.runner.ActivityReplyResponse = self.runner.ActivityReply(req, ctx)
        self.log("ACTIVITY REPLY RESPONSE:", resp)

    async def Activity(self, request: pb.handler.ActivityRequest):
        self.log("ACTIVITY", request)
        asyncio.ensure_future(self.call_execute(request.data))
        return pb.handler.ActivityResponse()

    async def ExecuteReply(self, request: pb.handler.ExecuteReplyRequest):
        self.log("EXECUTE", request)
        asyncio.ensure_future(self.call_activity_reply(request))
        return pb.handler.ExecuteReplyResponse()

    async def Done(self, request: pb.handler.DoneRequest):
        self.log("DONE", request)
        self.event.set()

    async def Log(self, request: pb.handler.LogRequest):
        self.log("LOG", request)
        return pb.handler.LogResponse()

    async def Print(self, request: pb.handler.PrintRequest):
        self.log("PRINT", request.message)
        return pb.handler.PrintResponse()

    async def Sleep(self, request: pb.handler.SleepRequest):
        self.log("SLEEP", request.duration_ms)
        await asyncio.sleep(request.duration_ms * 1000)
        return pb.handler.SleepResponse()

    async def Subscribe(self, request: pb.handler.SubscribeRequest):
        self.log("SUBSCRIBE", request)
        return pb.handler.SubscribeResponse(signal_id=next_signal_id())

    async def NextEvent(self, request: pb.handler.NextEventRequest):
        self.log("NEXT_EVENT", request)
        # TODO: Allow user to set events
        return pb.handler.NextEventResponse(
            event=pb.user_code.Event(data=b"next_event")
        )

    async def Unsubscribe(self, request: pb.handler.UnsubscribeRequest):
        self.log("UNSUBSCRIBE", request)
        return pb.handler.UnsubscribeResponse()

    async def StartSession(self, request: pb.handler.StartSessionRequest):
        self.log("START_SESSION", request)
        return pb.handler.StartSessionResponse()

    async def Signal(self, request: pb.handler.SignalRequest):
        self.log("SIGNAL", request)
        return pb.handler.SignalResponse()

    async def NextSignal(self, request: pb.handler.NextSignalRequest):
        self.log("NEXT_SIGNAL", request)
        return pb.handler.NextSignalResponse(
            signal=pb.handler.Signal(
                name="signal",
                value=pb.values.Nothing(),
            ),
        )

    async def EncodeJWT(self, request: pb.handler.EncodeJWTRequest):
        self.log("ENCODE_JWT", request)
        return pb.handler.EncodeJWTResponse()

    async def RefreshOAuthToken(self, request: pb.handler.RefreshRequest):
        self.log("REFRESH_OAUTH_TOKEN", request)
        return pb.handler.RefreshResponse(token="token")

    async def Health(self, request: pb.handler.HandlerHealthRequest):
        self.log("HEALTH", request)
        return pb.handler.HandlerHealthResponse()

    async def IsActivateRunner(self, request: pb.handler.IsActiveRunnerRequest):
        self.log("IS_ACTIVE_RUNNER", request)
        return pb.handler.IsActiveRunnerResponse(is_active=True)

    def log(self, func, msg):
        if self.verbose:
            # We can't use 'print' since main replaces it with a call to the worker Print
            sys.stdout.write(f"<<{func}>> {msg}\n")
        self.calls[func] += 1
