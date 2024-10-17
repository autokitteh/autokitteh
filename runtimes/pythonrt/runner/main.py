from base64 import b64decode
import builtins
from concurrent.futures import Future, ThreadPoolExecutor
from io import StringIO
import json
from multiprocessing import cpu_count
import os
from pathlib import Path
import pickle
import sys
from threading import Lock, Thread
from time import sleep
from traceback import TracebackException, print_exception

import grpc
from grpc_reflection.v1alpha import reflection

import loader
import log
import pb.autokitteh.remote.v1.remote_pb2 as pb
import pb.autokitteh.remote.v1.remote_pb2_grpc as rpc
from autokitteh import AttrDict, connections
from call import AKCall, full_func_name
from syscalls import SysCalls

SERVER_GRACE_TIMEOUT = 3  # seconds


class ActivityError(Exception):
    pass


def parse_entry_point(entry_point):
    """
    >>> parse_entry_point('review.py:on_github_pull_request')
    ('review', 'on_github_pull_request')
    """
    if ":" not in entry_point:
        raise ValueError(f"{entry_point!r} - missing :")

    file_name, func_name = entry_point.split(":", 1)
    if not file_name.endswith(".py"):
        raise ValueError(f"{entry_point!r} - not a Python file")

    return file_name[:-3], func_name


def exc_traceback(err):
    """Format traceback to JSONable list."""
    te = TracebackException.from_exception(err)
    return [
        pb.Frame(
            filename=frame.filename,
            lineno=frame.lineno,
            code=frame.line,
            name=frame.name,
        )
        for frame in te.stack
    ]


def display_err(fn, err):
    func_name = full_func_name(fn)
    log.exception("calling %s: %s", func_name, err)
    # Print the error to stderr so it'll show in session logs
    print(f"error: {err}", file=sys.stderr)
    print_exception(err, file=sys.stderr)


# Go passes HTTP event.data.body.bytes as base64 encode string
def fix_http_body(event):
    data = event.get("data")
    if not isinstance(data, dict):
        return

    body = data.get("body")
    if not isinstance(body, dict):
        return

    payload = body.get("bytes")
    if isinstance(payload, str):
        try:
            body["bytes"] = b64decode(payload)
        except ValueError:
            pass


def killIfStartWasntCalled(runner):
    if not runner.did_start:
        print("Start was not called, killing self")
        os._exit(1)


class Runner(rpc.RunnerServicer):
    def __init__(self, id, worker, code_dir, server):
        self.id = id
        self.worker: rpc.WorkerStub = worker
        self.code_dir = code_dir
        self.server: grpc.Server = server

        self.executor = ThreadPoolExecutor()

        self.lock = Lock()
        self.calls = {}  # id -> (fn, args, kw)
        self.replies = {}  # id -> future
        self._next_id = 0
        self._orig_print = print
        self._start_called = False

    def Exports(self, request: pb.ExportsRequest, context: grpc.ServicerContext):
        if request.file_name == "":
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "missing file name",
            )

        try:
            exports = list(loader.exports(self.code_dir, request.file_name))
        except OSError as err:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(err))

        return pb.ExportsResponse(exports=exports)

    def should_keep_running(self, initial_delay=10, period=10):
        sleep(initial_delay)
        if not self._start_called:
            log.error("Start not called after %dsec", initial_delay)
            self.server.stop(SERVER_GRACE_TIMEOUT)
            return

        # Check that we are still active
        while True:
            try:
                req = pb.IsActiveRunnerRequest(runner_id=self.id)
                res = self.worker.IsActiveRunner(req)
                if res.error:
                    break
            except grpc.RpcError:
                break
            sleep(period)

        log.error("could not verify if should keep running, killing self")
        self.server.stop(SERVER_GRACE_TIMEOUT)

    def Start(self, request: pb.StartRequest, context: grpc.ServicerContext):
        if self._start_called:
            log.error("already called start before")
            return pb.StartResponse(error="start already called")

        self._start_called = True
        log.info("start request: %r", request)

        self.syscalls = SysCalls(self.id, self.worker)
        mod_name, fn_name = parse_entry_point(request.entry_point)

        # Monkey patch some functions, should come before we import user code.
        builtins.print = self.ak_print
        connections.encode_jwt = self.syscalls.ak_encode_jwt
        connections.refresh_oauth = self.syscalls.ak_refresh_oauth

        call = AKCall(self)
        mod = loader.load_code(self.code_dir, call, mod_name)
        call.set_module(mod)

        fn = getattr(mod, fn_name, None)
        if not callable(fn):
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"function {fn_name!r} not found",
            )

        event = json.loads(request.event.data)

        fix_http_body(event)
        event = AttrDict(event)
        self.executor.submit(self.on_event, fn, event)

        return pb.StartResponse()

    def Execute(self, request: pb.ExecuteRequest, context: grpc.ServicerContext):
        call_id = request.data.decode()
        with self.lock:
            call_info = self.calls.pop(call_id, None)

        if call_info is None:
            error = f"call_id {call_id!r} not found"
            with self.lock:
                fut = self.replies.pop(call_id, None)
                if fut:
                    fut.set_exception(ActivityError(error))

            context.abort(grpc.StatusCode.INVALID_ARGUMENT, error)

        fn, args, kw = call_info
        log.info("calling %s, args=%r, kw=%r", full_func_name(fn), args, kw)
        result = err = None
        try:
            result = fn(*args, **kw)
        except Exception as e:
            display_err(fn, e)
            err = e

        resp = pb.ExecuteResponse(
            result=pickle.dumps(result, protocol=0),
        )

        if err:
            resp.error = str(err)
            tb = exc_traceback(err)
            resp.traceback.extend(tb)

            context.set_code(grpc.StatusCode.ABORTED)
            context.set_details(resp.error)

        return resp

    def ActivityReply(
        self, request: pb.ActivityReplyRequest, context: grpc.ServicerContext
    ):
        call_id = request.data.decode()
        with self.lock:
            fut = self.replies.pop(call_id, None)

        if fut is None:
            log.error("call_id %r not found", call_id)
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT, "call_id {call_id!r} not found"
            )

        try:
            result = pickle.loads(request.result)
        except Exception as err:
            log.exception(f"call_id {call_id!r}: result pickle: {err}")
            fut.set_exception(ActivityError(err))
            context.abort(
                grpc.StatusCode.INTERNAL, f"call_id {call_id!r}: result pickle: {err}"
            )

        fut.set_result(result)
        return pb.ActivityReplyResponse()

    def Health(self, request: pb.HealthRequest, context: grpc.ServicerContext):
        return pb.HealthResponse()

    def call_in_activity(self, fn, args, kw):
        fut = self.start_activity(fn, args, kw)
        return fut.result()

    def start_activity(self, fn, args, kw) -> Future:
        fn_name = full_func_name(fn)
        log.info("calling %s, args=%r, kw=%r", fn_name, args, kw)
        call_id = self.next_call_id()
        log.info("call_id %r", call_id)
        with self.lock:
            self.replies[call_id] = fut = Future()
            self.calls[call_id] = (fn, args, kw)

        req = pb.ActivityRequest(
            runner_id=self.id,
            data=call_id.encode(),
            call_info=pb.CallInfo(
                function=fn.__name__,  # AK rejects json.loads
                args=[repr(a) for a in args],
                kwargs={k: repr(v) for k, v in kw.items()},
            ),
        )
        log.info("activity: sending %r", req)
        resp = self.worker.Activity(req)
        if resp.error:
            raise ActivityError(resp.error)
        log.info("activity request ended")
        return fut

    def on_event(self, fn, event):
        log.info("on_event: start: %r", event)

        # TODO: This is similar to Execute, merge?
        err = result = None
        try:
            result = fn(event)
        except Exception as e:
            display_err(fn, e)
            err = e

        log.info("on_event: end: result=%r, err=%r", result, err)
        req = pb.DoneRequest(
            runner_id=self.id,
            # TODO: We want value that AK can understand (proto.Value)
            result=pickle.dumps(result, protocol=0),
        )

        if err:
            req.error = str(err)
            tb = exc_traceback(err)
            req.traceback.extend(tb)

        resp = self.worker.Done(req)
        if resp.Error:
            log.error("on_event: done error: %r", resp.error)

    def syscall(self, fn, args, kw):
        return self.syscalls.call(fn, args, kw)

    def next_call_id(self) -> str:
        with self.lock:
            self._next_id += 1
            return f"call_id_{self._next_id:03d}"

    def ak_print(self, *objects, sep=" ", end="\n", file=None, flush=False):
        io = StringIO()
        self._orig_print(*objects, sep=sep, end=end, flush=flush, file=io)
        text = io.getvalue()
        self._orig_print(text, file=file)  # Print also to original destination

        req = pb.PrintRequest(
            runner_id=self.id,
            message=text,
        )

        try:
            self.worker.Print(req)
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.UNAVAILABLE or grpc.StatusCode.CANCELLED:
                log.error("grpc canclled or unavailable, killing self")
                self.server.stop(SERVER_GRACE_TIMEOUT)
            log.error("print: %s", err)


def is_valid_port(port):
    return port >= 0 and port <= 65535


def validate_args(args):
    if not is_valid_port(args.port):
        raise ValueError(f"invalid port: {args.port!r}")

    if ":" not in args.worker_address:
        raise ValueError("worker address must be in the form host:port")
    host, port = args.worker_address.split(":")
    if host == "":
        raise ValueError(f"empty host in {args.worker_address!r}")

    port = int(port)
    if not is_valid_port(port):
        raise ValueError(f"invalid port in {args.worker_address!r}")

    if args.runner_id == "":
        raise ValueError("runner ID cannot be empty")


class LoggingInterceptor(grpc.ServerInterceptor):
    runner_id = None

    def intercept_service(self, continuation, handler_call_details):
        log.info("runner_id %s, call %s", self.runner_id, handler_call_details.method)
        return continuation(handler_call_details)

    def __init__(self, runner_id) -> None:
        self.runner_id = runner_id
        super().__init__()


def dir_type(value):
    path = Path(value)
    if not path.is_dir():
        raise ValueError(f"{value!r} is not a directory")

    return path


if __name__ == "__main__":
    from argparse import ArgumentParser

    parser = ArgumentParser(description="Python runner")
    parser.add_argument(
        "--worker-address", help="Worker address (host:port)", default="localhost:9292"
    )
    parser.add_argument(
        "--skip-check-worker",
        help="do not check connection to worker on startup",
        action="store_true",
    )
    parser.add_argument("--port", help="port to listen on", default=9293, type=int)
    parser.add_argument("--runner-id", help="runner ID", default="runner-1")
    parser.add_argument(
        "--code-dir",
        help="directory of user code",
        default="/workflow",
        type=dir_type,
    )
    args = parser.parse_args()

    try:
        validate_args(args)
    except ValueError as err:
        raise SystemExit(f"error: {err}")

    # Support importing local files
    sys.path.append(str(args.code_dir))

    chan = grpc.insecure_channel(args.worker_address)
    worker = rpc.WorkerStub(chan)
    if not args.skip_check_worker:
        req = pb.HealthRequest()
        try:
            resp = worker.Health(req)
        except grpc.RpcError as err:
            raise SystemExit(f"error: worker not available - {err}")

    log.info("connected to worker at %r", args.worker_address)

    server = grpc.server(
        thread_pool=ThreadPoolExecutor(max_workers=cpu_count() * 8),
        interceptors=[LoggingInterceptor(args.runner_id)],
    )
    runner = Runner(args.runner_id, worker, args.code_dir, server)
    rpc.add_RunnerServicer_to_server(runner, server)
    services = (
        pb.DESCRIPTOR.services_by_name["Runner"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(services, server)

    server.add_insecure_port(f"[::]:{args.port}")
    server.start()
    log.info("server running on port %d", args.port)

    if not args.skip_check_worker:
        Thread(target=runner.should_keep_running, daemon=True).start()
    log.info("setup should keep running thread")

    server.wait_for_termination()
