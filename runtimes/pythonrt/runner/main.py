import builtins
import inspect
import json
import pickle
import sys
from base64 import b64decode
from collections import namedtuple
from concurrent.futures import Future, ThreadPoolExecutor
from io import StringIO
from multiprocessing import cpu_count
from pathlib import Path
from threading import Lock, Thread, Timer
from time import sleep
from traceback import TracebackException, format_exception

import autokitteh
import grpc
import loader
import log
import pb
import values

# from audit import make_audit_hook  # TODO(ENG-1893): uncomment this.
from autokitteh import AttrDict, connections
from autokitteh.errors import AutoKittehError
from call import AKCall, full_func_name, is_marked_activity
from syscalls import SysCalls

# Timeouts are in seconds
SERVER_GRACE_TIMEOUT = 3
START_TIMEOUT = 10


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


def pb_traceback(tb):
    """Convert traceback to a list of pb.user_code.Frame for serialization."""
    return [
        pb.user_code.Frame(
            filename=frame.filename,
            lineno=frame.lineno,
            code=frame.line,
            name=frame.name,
        )
        for frame in tb.stack
    ]


pickle_help = """
=======================================================================================================
The below error means you need to use the @autokitteh.activity decorator.
See https://docs.autokitteh.com/develop/python/#function-arguments-and-return-values-must-be-pickleable
for more details.
=======================================================================================================
"""


def result_error(err):
    io = StringIO()

    if "pickle" in str(err):
        print(pickle_help, file=io)

    exc = "".join(format_exception(err))
    message = s if (s := str(err)) else repr(err)
    print(f"error: {message}\n\n{exc}", file=io)

    return io.getvalue()


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


def abort_with_exception(context, status, err):
    io = StringIO()
    for line in format_exception(err):
        io.write(line)
    text = io.getvalue()
    context.abort(status, text)


def set_exception_args(err):
    """Set exception "args" attribute to match __init__ signature so it can be un-pickled.

    See https://stackoverflow.com/questions/41808912/cannot-unpickle-exception-subclass
    """
    code = getattr(err.__init__, "__code__", None)
    if code is None:  # Built-in
        return

    init_args = inspect.getargs(code).args
    if not init_args:
        return

    if init_args[0] == "self":
        init_args = init_args[1:]

    err_args = getattr(err, "args")
    if len(err_args) == len(init_args):
        return

    extra = []
    for name in init_args[len(err_args) :]:
        extra.append(getattr(err, name, None))

    err.args += tuple(extra)


Call = namedtuple("Call", "fn args kw fut")
Result = namedtuple("Result", "value error traceback")


def is_pickleable(err):
    try:
        data = pickle.dumps(err)
        pickle.loads(data)
        return True
    except (TypeError, pickle.PickleError):
        return False


def restore_error(err):
    if isinstance(err, Exception):
        return err

    if not isinstance(err, tuple):
        raise TypeError("excepted a tuple, got %r", err)

    if len(err) < 3:
        raise ValueError("reduce tuple should be at least 3 elements, got %r", err)

    cls, _, state = err[:3]
    obj = cls.__new__(cls)
    obj.__setstate__(state)
    return obj


class Runner(pb.runner_rpc.RunnerService):
    def __init__(self, id, worker, code_dir, server):
        self.id = id
        self.worker: pb.handler_rpc.HandlerServiceStub = worker
        self.code_dir = code_dir
        self.server: grpc.Server = server

        self.executor = ThreadPoolExecutor()

        self.lock = Lock()
        self.activity_call = None
        self._orig_print = print
        self._start_called = False
        self._inactivty_timer = Timer(START_TIMEOUT, self.stop_if_start_not_called)
        self._inactivty_timer.start()

    def stop_if_start_not_called(self):
        log.error("Start not called after %s seconds, terminating", START_TIMEOUT)
        if self.server:
            self.server.stop(SERVER_GRACE_TIMEOUT)

    def Exports(self, request: pb.runner.ExportsRequest, context: grpc.ServicerContext):
        if request.file_name == "":
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "missing file name",
            )

        try:
            exports = list(loader.exports(self.code_dir, request.file_name))
        except OSError as err:
            abort_with_exception(context, grpc.StatusCode.INVALID_ARGUMENT, err)

        return pb.runner.ExportsResponse(exports=exports)

    def should_keep_running(self, initial_delay=10, period=10):
        sleep(initial_delay)
        if not self._start_called:
            log.error("Start not called after %dsec", initial_delay)
            self.server.stop(SERVER_GRACE_TIMEOUT)
            return

        # Check that we are still active
        while True:
            try:
                req = pb.handler.IsActiveRunnerRequest(runner_id=self.id)
                res = self.worker.IsActiveRunner(req)
                if res.error:
                    break
            except grpc.RpcError:
                break
            sleep(period)

        log.error("could not verify if should keep running, killing self")
        self.server.stop(SERVER_GRACE_TIMEOUT)

    def patch_ak_funcs(self):
        connections.encode_jwt = self.syscalls.ak_encode_jwt
        connections.refresh_oauth = self.syscalls.ak_refresh_oauth

        autokitteh.start = self.syscalls.ak_start
        autokitteh.next_event = self.syscalls.ak_next_event
        autokitteh.subscribe = self.syscalls.ak_subscribe
        autokitteh.unsubscribe = self.syscalls.ak_unsubscribe

        # Not ak, but patching print as well
        builtins.print = self.ak_print

    def Start(self, request: pb.runner.StartRequest, context: grpc.ServicerContext):
        if self._start_called:
            log.error("already called start before")
            return pb.runner.StartResponse(error="start already called")

        self._inactivty_timer.cancel()

        self._start_called = True
        log.info("start request: %r", request.entry_point)

        self.syscalls = SysCalls(self.id, self.worker, log)
        mod_name, fn_name = parse_entry_point(request.entry_point)

        # Must be before we load user code
        self.patch_ak_funcs()

        ak_call = AKCall(self, self.code_dir)
        try:
            mod = loader.load_code(self.code_dir, ak_call, mod_name)
        except Exception as err:
            self.ak_print(result_error(err))
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"can't load {mod_name} from {self.code_dir} - {err}",
            )

        ak_call.set_module(mod)

        fn = getattr(mod, fn_name, None)
        if not callable(fn):
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"function {fn_name!r} not found",
            )

        event = json.loads(request.event.data)

        fix_http_body(event)
        event = AttrDict(event)

        # TODO(ENG-1893): Disabled temporarily due to issues with HubSpot client - need to investigate.
        # # Warn on I/O outside an activity. Should come after importing the user module
        # hook = make_audit_hook(ak_call, self.code_dir)
        # sys.addaudithook(hook)

        if is_marked_activity(fn):
            orig_fn = fn

            def handler(event):
                return ak_call(orig_fn, event)

            fn = handler

        self.executor.submit(self.on_event, fn, event)

        return pb.runner.StartResponse()

    def Execute(self, request: pb.runner.ExecuteRequest, context: grpc.ServicerContext):
        with self.lock:
            call: Call = self.activity_call

        if call is None:
            context.abort(grpc.StatusCode.INTERNAL, "no pending activity calls")

        result = self._call(call.fn, call.args, call.kw)
        try:
            data = pickle.dumps(result)
        except (TypeError, pickle.PickleError) as err:
            # Print so it'll get to session log
            print(f"error: cannot pickle result - {err}")
            print(pickle_help)
            context.abort(grpc.StatusCode.INTERNAL, f"can't pickle result - {err}")

        resp = pb.runner.ExecuteResponse(
            result=pb.values.Value(
                custom=pb.values.Custom(
                    data=data,
                    value=values.safe_wrap(result.value),
                ),
            )
        )

        return resp

    def ActivityReply(
        self, request: pb.runner.ActivityReplyRequest, context: grpc.ServicerContext
    ):
        if request.error or not request.result.custom.data:
            error = request.error or "activity reply not a Custom value"
            req = pb.handler.DoneRequest(
                runner_id=self.id,
                error=error,
            )
            try:
                self.worker.Done(req)
            except grpc.RpcError as err:
                log.error("done send error: %r", err)
                self.server.stop(SERVER_GRACE_TIMEOUT)

            return pb.runner.ActivityReplyResponse(error=request.error)

        result = None
        try:
            result = pickle.loads(request.result.custom.data)
        except Exception as err:
            log.exception(f"can't decode data: pickle: {err}")
            abort_with_exception(context, grpc.StatusCode.INTERNAL, err)

        if not isinstance(result, Result):
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT, "ActivityReply data not a Result"
            )

        with self.lock:
            call = self.activity_call
            self.activity_call = None

        if call is None:
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT, "ActivityReply without pending calls"
            )

        if result.error:
            try:
                error = restore_error(result.error)
            except (TypeError, ValueError) as err:
                log.exception("can't restore error: %r", err)
                error = AutoKittehError(repr(result.error))
            call.fut.set_exception(error)
        else:
            call.fut.set_result(result.value)

        return pb.runner.ActivityReplyResponse()

    def Health(
        self,
        request: pb.runner.RunnerHealthRequest,
        context: grpc.ServicerContext,
    ):
        return pb.runner.RunnerHealthResponse()

    def call_in_activity(self, fn, args, kw):
        log.info("call_in_activity: %s", full_func_name(fn))
        fut = self.start_activity(fn, args, kw)
        return fut.result()

    def start_activity(self, fn, args, kw) -> Future:
        fn_name = full_func_name(fn)
        log.info("calling %s", fn_name)
        call = Call(fn, args, kw, Future())
        with self.lock:
            if self.activity_call:
                log.error("nested activity: %r < %r", self.activity_call, call)
                raise RuntimeError(f"nested activity: {self.activity_call} < {call}")
            self.activity_call = call

        req = pb.handler.ActivityRequest(
            runner_id=self.id,
            call_info=pb.handler.CallInfo(
                function=fn.__name__,  # AK rejects __qualname__ such as "json.loads"
                args=[values.safe_wrap(a) for a in args],
                kwargs={k: values.safe_wrap(v) for k, v in kw.items()},
            ),
        )
        log.info("activity: sending")
        resp = self.worker.Activity(req)
        if resp.error:
            raise ActivityError(resp.error)
        log.info("activity request ended")
        return call.fut

    def _call(self, fn, args, kw):
        func_name = full_func_name(fn)
        log.info("calling %s", func_name)
        value = error = tb = None
        try:
            value = fn(*args, **kw)
            if isinstance(value, Exception):
                set_exception_args(value)
        except Exception as err:
            log.error("%s raised: %s", func_name, err)
            tb = TracebackException.from_exception(err)
            error = err
            set_exception_args(error)

        if not is_pickleable(error):
            log.info("non pickleable: %r", error)
            error = error.__reduce__()

        return Result(value, error, tb)

    def on_event(self, fn, event):
        func_name = full_func_name(fn)
        log.info("start event: %s", func_name)

        result = self._call(fn, [event], {})

        log.info("event end: error=%r", result.error)
        req = pb.handler.DoneRequest(
            runner_id=self.id,
        )

        if result.error:
            req.error = result_error(result.error)
            tb = pb_traceback(result.traceback)
            req.traceback.extend(tb)
        else:
            try:
                data = pickle.dumps(result)
                req.result.custom.data = data
                req.result.custom.value.CopyFrom(values.safe_wrap(result.value))
            except (TypeError, pickle.PickleError) as err:
                req.error = f"can't pickle {result.value} - {err}"

        try:
            self.worker.Done(req)
        except grpc.RpcError as err:
            log.error("on_event: done send error: %r", err)

    def ak_print(self, *objects, sep=" ", end="\n", file=None, flush=False):
        io = StringIO()
        self._orig_print(*objects, sep=sep, end=end, flush=flush, file=io)
        text = io.getvalue()
        self._orig_print(text, file=file)  # Print also to original destination

        req = pb.handler.PrintRequest(
            runner_id=self.id,
            message=text,
        )

        try:
            self.worker.Print(req)
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.UNAVAILABLE or grpc.StatusCode.CANCELLED:
                log.error("grpc cancelled or unavailable, killing self")
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
    worker = pb.handler_rpc.HandlerServiceStub(chan)
    if not args.skip_check_worker:
        req = pb.handler.HandlerHealthRequest()
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
    # rpc.add_RunnerServicer_to_server(runner, server)
    pb.runner_rpc.add_RunnerServiceServicer_to_server(runner, server)

    server.add_insecure_port(f"[::]:{args.port}")
    server.start()
    log.info("server running on port %d", args.port)

    if not args.skip_check_worker:
        Thread(target=runner.should_keep_running, daemon=True).start()
        log.info("started 'should_keep_running' thread")

    server.wait_for_termination()
