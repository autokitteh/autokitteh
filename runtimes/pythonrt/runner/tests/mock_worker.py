"""A mock worker so we can debug the runner without ak."""

import sys
from itertools import count
from threading import Event, Thread
from time import sleep
from unittest.mock import MagicMock

import pb

_next_id = count(1).__next__


def next_signal_id():
    return f"signal-{_next_id()}"


# We can't use 'print' since main replaces it with a call to the worker Print
def log(func, msg):
    sys.stdout.write(f"<<{func}>> {msg}\n")


class MockWorker(pb.handler_rpc.HandlerService):
    def __init__(self, runner: pb.runner_rpc.RunnerService):
        self.runner = runner
        self.runner.worker = self
        self.event = Event()

    def start(self, entry_point, data):
        req = pb.runner.StartRequest(
            entry_point=entry_point, event=pb.user_code.Event(data=data)
        )
        ctx = MagicMock()
        resp: pb.runner.StartResponse = self.runner.Start(req, ctx)
        log("START RESPONSE:", resp)
        self.event.wait()

    def call_execute(self, data):
        ctx = MagicMock()

        sleep(0.1)
        req = pb.runner.ExecuteRequest(data=data)
        resp: pb.runner.ExecuteResponse = self.runner.Execute(req, ctx)
        log("EXECUTE RESPONSE:", resp)

    def call_activity_reply(self, msg: pb.handler.ExecuteReplyRequest):
        ctx = MagicMock()

        req = pb.runner.ActivityReplyRequest(
            result=msg.result,
            error=msg.error,
            # TODO: Traceback?
        )
        resp: pb.runner.ActivityReplyResponse = self.runner.ActivityReply(req, ctx)
        log("ACTIVITY REPLY RESPONSE:", resp)

    def Activity(self, request: pb.handler.ActivityRequest):
        log("ACTIVITY", request)
        Thread(
            target=self.call_execute,
            args=(request.data,),
            daemon=True,
        ).start()
        return pb.handler.ActivityResponse()

    def ExecuteReply(self, request: pb.handler.ExecuteReplyRequest):
        log("EXECUTE", request)
        Thread(
            target=self.call_activity_reply,
            args=(request,),
            daemon=True,
        ).start()
        return pb.handler.ExecuteReplyResponse()

    def Done(self, request: pb.handler.DoneRequest):
        log("DONE", request)
        self.event.set()

    def Log(self, request: pb.handler.LogRequest):
        log("LOG", request)
        return pb.handler.LogResponse()

    def Print(self, request: pb.handler.PrintRequest):
        log("PRINT", request.message)
        return pb.handler.PrintResponse()

    def Sleep(self, request: pb.handler.SleepRequest):
        log("SLEEP", request.duration)
        sleep(request.duration_ms * 1000)
        return pb.handler.SleepResponse()

    def Subscribe(self, request: pb.handler.SubscribeRequest):
        log("SUBSCRIBE", request)
        return pb.handler.SubscribeResponse(signal_id=next_signal_id())

    def NextEvent(self, request: pb.handler.NextEventRequest):
        log("NEXT_EVENT", request)
        # TODO: Allow user to set events
        return pb.handler.NextEventResponse(
            event=pb.user_code.Event(data=b"next_event")
        )

    def Unsubscribe(self, request: pb.handler.UnsubscribeRequest):
        log("UNSUBSCRIBE", request)
        return pb.handler.UnsubscribeResponse()

    def StartSession(self, request: pb.handler.StartSessionRequest):
        log("START_SESSION", request)
        return pb.handler.StartSessionResponse()

    def Signal(self, request: pb.handler.SignalRequest):
        log("SIGNAL", request)
        return pb.handler.SignalResponse()

    def NextSignal(self, request: pb.handler.NextSignalRequest):
        log("NEXT_SIGNAL", request)
        return pb.handler.NextSignalResponse(
            signal=pb.handler.Signal(
                name="signal",
                value=pb.values.Nothing(),
            ),
        )

    def EncodeJWT(self, request: pb.handler.EncodeJWTRequest):
        log("ENCODE_JWT", request)
        return pb.handler.EncodeJWTResponse()

    def RefreshOAuthToken(self, request: pb.handler.RefreshRequest):
        log("REFRESH_OAUTH_TOKEN", request)
        return pb.handler.RefreshResponse(token="token")

    def Health(self, request: pb.handler.HandlerHealthRequest):
        log("HEALTH", request)
        return pb.handler.HandlerHealthResponse()

    def IsActivateRunner(self, request: pb.handler.IsActiveRunnerRequest):
        log("IA_ACTIVE_RUNNER", request)
        return pb.handler.IsActiveRunnerResponse(is_active=True)
