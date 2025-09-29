import asyncio
import builtins
import inspect
import json
import os
import pickle
import sys
from base64 import b64decode
from collections import namedtuple
from concurrent.futures import Future, ThreadPoolExecutor
from functools import update_wrapper
from io import StringIO
from multiprocessing import cpu_count
from pathlib import Path
from threading import Lock, Thread, Timer
from time import sleep
from traceback import TracebackException, format_exception

import autokitteh
import autokitteh.store
import grpc
import loader
import log
import pb
import values

# from audit import make_audit_hook  # TODO(ENG-1893): uncomment this.
from autokitteh import AttrDict, Event, connections
from autokitteh.errors import AutoKittehError
from call import AKCall, activity_marker, full_func_name
from syscalls import SysCalls, mark_no_activity

# Timeouts are in seconds
SERVER_GRACE_TIMEOUT = 3
DEFAULT_START_TIMEOUT = 10


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


# This is the same as pb.user_code.Frame, but we want to decouple internal code from API
# proto definition.
Frame = namedtuple("Frame", "filename lineno code name")


def tb_stack(tb):
    return [
        Frame(
            filename=frame.filename,
            lineno=frame.lineno,
            code=frame.line,
            name=frame.name,
        )
        for frame in tb.stack
    ]


def pb_traceback(stack):
    """Convert traceback to a list of pb.user_code.Frame for serialization."""
    return [
        pb.user_code.Frame(
            filename=frame.filename,
            lineno=frame.lineno,
            code=frame.code,
            name=frame.name,
        )
        for frame in stack
    ]


def filter_traceback(stack, user_code):
    """Filter out first part of traceback until first user code frame."""
    for i, frame in enumerate(stack):
        try:
            if Path(frame.filename).is_relative_to(user_code):
                return stack[i:]
        except (ValueError, OSError) as err:
            log.error(f"filter stack: {err}")
            return stack

    return stack


pickle_help = """
=======================================================================================================
The below error means you need to use the @autokitteh.activity decorator.
See https://docs.autokitteh.com/develop/python/#function-arguments-and-return-values-must-be-pickleable
for more details.
=======================================================================================================
"""


suggest_add_package = """
=======================================================================================================
The below error means you need to add a package to the Python environment.
Web platform: Create a requirements.txt file and add module to the requirements.txt file.
Self hosted: add module to AutoKitteh virtual environment. 
See https://docs.autokitteh.com/develop/python#installing-python-packages for more details.
=======================================================================================================
"""


# Go passes HTTP event.data.body.bytes as base64 encode string
def fix_http_body(inputs):
    data = inputs.get("data")
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


def abort_with_exception(context, status, err, show_pickle_help=False):
    io = StringIO()
    if show_pickle_help:
        print(pickle_help, file=io)
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

    try:
        err.args += tuple(extra)
    except Exception as e:
        # Some errors (msgraph ODataError) raise when you set args
        log.warning("can't set args for %r - %r", err, e)


Call = namedtuple("Call", "fn args kw fut")
Result = namedtuple("Result", "value error traceback")


def is_pickleable(err):
    try:
        data = pickle.dumps(err)
        pickle.loads(data)
        return True
    except (TypeError, pickle.PickleError):
        return False
    except Exception as pickle_err:
        # This is unexpected, but we can't not handle it.
        # Logging so we can investigate.
        log.exception("unexpected error: %r", pickle_err)
        tb = "".join(format_exception(pickle_err))
        log.error("traceback:\n%r", tb)
        log.error("error we tried to pickle: %r", err)
        try:
            attrs = vars(err)
            log.error("exception attributes: %r", attrs)
        except Exception:
            pass
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


def short_name(func_name: str):
    """
    >>> short_name('pickle.dumps')
    'dumps'
    >>> short_name('urlopen')
    'urlopen'
    """
    i = func_name.rfind(".")
    if i == -1:
        return func_name

    return func_name[i + 1 :]


def force_close(server):
    server.stop(SERVER_GRACE_TIMEOUT)
    os._exit(1)


MAX_SIZE_OF_REQUEST = 1 * 1024 * 1024


def format_size(size):
    if size < 1024:
        return f"{size} B"
    elif size < 1024 * 1024:
        return f"{size // 1024} KB"
    else:
        return f"{size // (1024 * 1024)} MB"


class Runner(pb.runner_rpc.RunnerService):
    def __init__(
        self, id, worker, code_dir, server, start_timeout=DEFAULT_START_TIMEOUT
    ):
        self.id = id
        self.worker: pb.handler_rpc.HandlerServiceStub = worker
        self.code_dir = code_dir
        self.server: grpc.Server = server

        self.executor = ThreadPoolExecutor()

        self.lock = Lock()
        self.activity_call: Call = None
        self._orig_print = print
        self._start_called = False
        self._inactivity_timer = Timer(
            start_timeout, self.stop_if_start_not_called, args=(start_timeout,)
        )
        self._inactivity_timer.start()
        self._stopped = False

    def result_error(self, err):
        io = StringIO()

        if "No module named" in str(err):
            self._orig_print(suggest_add_package, file=io)

        if "pickle" in str(err):
            self._orig_print(pickle_help, file=io)

        exc = "".join(format_exception(err))
        self._orig_print(f"error: {err!r}\n\n{exc}", file=io)

        return io.getvalue()

    def stop_if_start_not_called(self, timeout):
        log.error("Start not called after %s seconds, terminating", timeout)
        if self.server:
            force_close(self.server)

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
            force_close(self.server)
            return

        # Check that we are still active
        while not self._stopped:
            try:
                req = pb.handler.IsActiveRunnerRequest(runner_id=self.id)
                res = self.worker.IsActiveRunner(req)
                if res.error:
                    break
            except grpc.RpcError:
                break
            sleep(period)

        if self._stopped:
            log.info("Runner %s stopped, stopping should_keep_running loop", self.id)
            return

        log.error("could not verify if should keep running, killing self")
        force_close(self.server)

    def patch_ak_funcs(self):
        connections.encode_jwt = self.syscalls.ak_encode_jwt
        connections.refresh_oauth = self.syscalls.ak_refresh_oauth

        autokitteh.del_value = self.syscalls.ak_del_value
        autokitteh.get_value = self.syscalls.ak_get_value
        autokitteh.list_values_keys = self.syscalls.ak_list_values_keys
        autokitteh.mutate_value = self.syscalls.ak_mutate_value
        # Need to patch autokitteh.store as well for the Store API
        autokitteh.store.del_value = self.syscalls.ak_del_value
        autokitteh.store.get_value = self.syscalls.ak_get_value
        autokitteh.store.list_values_keys = self.syscalls.ak_list_values_keys
        autokitteh.store.mutate_value = self.syscalls.ak_mutate_value

        autokitteh.next_event = self.syscalls.ak_next_event
        autokitteh.next_signal = self.syscalls.ak_next_signal
        autokitteh.set_value = self.syscalls.ak_set_value
        autokitteh.add_values = self.syscalls.ak_add_values
        autokitteh.signal = self.syscalls.ak_signal
        autokitteh.start = self.syscalls.ak_start
        autokitteh.subscribe = self.syscalls.ak_subscribe
        autokitteh.unsubscribe = self.syscalls.ak_unsubscribe
        autokitteh.outcome = self.syscalls.ak_outcome
        autokitteh.http_outcome = self.syscalls.ak_http_outcome

        # Not ak, but patching print as well
        builtins.print = self.ak_print

    def Start(self, request: pb.runner.StartRequest, context: grpc.ServicerContext):
        # NOTE: Don't do any prints here, ak is not ready for them yet.
        if self._start_called:
            log.error("already called start before")
            return pb.runner.StartResponse(error="start already called")

        self._inactivity_timer.cancel()

        self._start_called = True
        log.info("start request: %r", request.entry_point)

        self.syscalls = SysCalls(self.id, self.worker, log)
        mod_name, fn_name = parse_entry_point(request.entry_point)

        inputs = json.loads(request.event.data)

        fix_http_body(inputs)

        event = Event(
            data=AttrDict(inputs.get("data", {})),
            session_id=inputs.get("session_id"),
        )

        # Must be before we load user code
        self.patch_ak_funcs()

        ak_call = AKCall(self, self.code_dir) if request.is_durable else None

        try:
            mod = loader.load_code(self.code_dir, ak_call, mod_name)
        except Exception as err:
            # Can't use ak_print here - ak not ready yet.
            Thread(
                target=self.server.stop, args=(SERVER_GRACE_TIMEOUT,), daemon=True
            ).start()
            err_text = self.result_error(err)
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"can't load {mod_name} from {self.code_dir} - {err_text}",
            )

        fn = getattr(mod, fn_name, None)
        if not callable(fn):
            Thread(
                target=self.server.stop, args=(SERVER_GRACE_TIMEOUT,), daemon=True
            ).start()
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"function {fn_name!r} not found",
            )

        if ak_call:
            ak_call.set_module(mod)

            # TODO(ENG-1893): Disabled temporarily due to issues with HubSpot client - need to investigate.
            # # Warn on I/O outside an activity. Should come after importing the user module
            # hook = make_audit_hook(ak_call, self.code_dir)
            # sys.addaudithook(hook)

            # Top-level handler marked as activity.
            if activity_marker(fn):
                orig_fn = fn

                def handler(event):
                    return ak_call(orig_fn, event)

                update_wrapper(handler, orig_fn)
                fn = handler

        self.executor.submit(self.on_event, fn, event)

        return pb.runner.StartResponse()

    def execute(self, fn, args, kw):
        req = pb.handler.ExecuteReplyRequest(
            runner_id=self.id,
        )

        result = self._call(fn, args, kw)
        try:
            data = pickle.dumps(result)
            req.result.custom.data = data
            req.result.custom.value.CopyFrom(values.safe_wrap(result.value))
            size_of_request = req.ByteSize()
            if size_of_request > MAX_SIZE_OF_REQUEST:
                # reset the req.result and use error only
                req = pb.handler.ExecuteReplyRequest(
                    runner_id=self.id,
                )
                req.error = "response size too large"
                print(
                    f"response size {format_size(size_of_request)} is too large, max allowed is {format_size(MAX_SIZE_OF_REQUEST)}"
                )
        except Exception as err:
            # Print so it'll get to session log
            msg = f"error processing result - {err!r}"
            print(f"error: {msg}")
            print(self.result_error(err))
            req.error = msg

        try:
            log.info("execute reply")
            resp = self.worker.ExecuteReply(req)
            if resp.error:
                log.error("execute reply: %r", resp.error)
                # TODO: need to handle this case (ENG-2253)
            return
        except grpc.RpcError as err:
            log.error("execute reply send error: %r", err)
            # TODO: need to handle this case (ENG-2253)

        # for now, if we got here, we need to kill self
        # should handle better with ENG-2253
        force_close(self.server)

    def Execute(self, request: pb.runner.ExecuteRequest, context: grpc.ServicerContext):
        with self.lock:
            call: Call = self.activity_call

        if call is None:
            context.abort(grpc.StatusCode.INTERNAL, "no pending activity calls")

        self.executor.submit(self.execute, call.fn, call.args, call.kw)
        return pb.runner.ExecuteResponse()

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
                force_close(self.server)

            return pb.runner.ActivityReplyResponse(error=request.error)

        result = None
        try:
            result = pickle.loads(request.result.custom.data)
        except Exception as err:
            log.exception(f"can't decode data: pickle: {err}")
            abort_with_exception(
                context, grpc.StatusCode.INTERNAL, err, show_pickle_help=True
            )

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
            if inspect.iscoroutinefunction(call.fn):
                # Wrap async result from activity so it can be awaited
                async def future():
                    return result.value

                value = future()
            else:
                value = result.value

            call.fut.set_result(value)

        return pb.runner.ActivityReplyResponse()

    def Health(
        self,
        request: pb.runner.RunnerHealthRequest,
        context: grpc.ServicerContext,
    ):
        duration = monotonic() - start_time
        log.info("health check (duration = %.2fsec)", duration)

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
                # AK rejects __qualname__ such as "json.loads"
                function=short_name(fn_name),
                args=[values.safe_wrap(a) for a in args],
                kwargs={k: values.safe_wrap(v) for k, v in kw.items()},
            ),
        )
        req_size = req.ByteSize()
        if req_size > MAX_SIZE_OF_REQUEST:
            raise ActivityError(
                f"Request payload size {format_size(req_size)} is larger than maximum supported ({format_size(MAX_SIZE_OF_REQUEST)})."
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
        value = error = stack = None
        try:
            value = fn(*args, **kw)
            if asyncio.iscoroutine(value):
                value = asyncio.run(value)
        except BaseException as err:
            log.error("%s raised: %r", func_name, err)
            tb = TracebackException.from_exception(err)
            # In some cases, tb can contains code objects that are not pickleable
            stack = tb_stack(tb)
            error = err
            set_exception_args(error)
            # In asyncio value will be a coroutine which can't be pickled
            value = None

        if isinstance(value, Exception):
            set_exception_args(value)

        if not is_pickleable(error):
            log.warning("non pickleable: %r", error)
            error = error.__reduce__()

        return Result(value, error, stack)

    def on_event(self, fn, event):
        func_name = full_func_name(fn)
        log.info("start event: %s", func_name)

        result = self._call(fn, [event], {})

        log.info("event end: error=%r", result.error)
        self._stopped = True
        req = pb.handler.DoneRequest(
            runner_id=self.id,
        )

        try:
            if result.error:
                error = restore_error(result.error)
                req.error = self.result_error(error)
                stack = filter_traceback(result.traceback, self.code_dir)
                tb = pb_traceback(stack)
                req.traceback.extend(tb)
            else:
                data = pickle.dumps(result)
                req.result.custom.data = data
                req.result.custom.value.CopyFrom(values.safe_wrap(result.value))
        except (TypeError, pickle.PickleError) as err:
            req.error = f"can't pickle {result.value} - {err}"
        except Exception as err:
            req.error = f"unexpected error: {err}"
            log.exception("on_event: %r", err)

        try:
            self.worker.Done(req)
        except Exception as err:
            log.error("on_event: done send error: %r", err)

        force_close(self.server)

    @mark_no_activity
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
            if err.code() in (grpc.StatusCode.UNAVAILABLE, grpc.StatusCode.CANCELLED):
                log.error("grpc cancelled or unavailable, killing self")
                force_close(self.server)
            log.error("print: %s", err)


def is_valid_port(port):
    return port >= 0 and port <= 65535


def validate_args(args):
    if not is_valid_port(args.port):
        raise ValueError(f"invalid port: {args.port!r}")

    if ":" not in args.worker_address:
        raise ValueError("worker address must be in the form host:port")
    host, port = args.worker_address.split(":", 1)
    if host == "":
        raise ValueError(f"empty host in {args.worker_address!r}")

    try:
        port = int(port)
    except ValueError as err:
        raise ValueError(f"{port!r}: bad port - {err}") from None

    if not is_valid_port(port):
        raise ValueError(f"invalid port in {args.worker_address!r}")

    if args.runner_id == "":
        raise ValueError("runner ID cannot be empty")

    if args.start_timeout <= 0:
        raise ValueError("start timeout must be positive")


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
    from time import monotonic

    # TODO(ENG-2089): Remove when we add telemetry.
    start_time = monotonic()

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
    parser.add_argument(
        "--start-timeout",
        help="timeout in seconds for start to be called",
        default=DEFAULT_START_TIMEOUT,
        type=int,
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
            resp = worker.Health(req, timeout=3)
        except grpc.RpcError as err:
            raise SystemExit(f"error: worker not available - {err}")

    duration = monotonic() - start_time
    log.info(
        "connected to worker at %r (duration = %.2fsec)", args.worker_address, duration
    )

    server = grpc.server(
        thread_pool=ThreadPoolExecutor(max_workers=cpu_count() * 8),
        interceptors=[LoggingInterceptor(args.runner_id)],
    )
    runner = Runner(args.runner_id, worker, args.code_dir, server, args.start_timeout)
    pb.runner_rpc.add_RunnerServiceServicer_to_server(runner, server)

    server.add_insecure_port(f"[::]:{args.port}")
    server.start()

    log.info("server running on port %d (duration = %.2fsec)", args.port, duration)

    if not args.skip_check_worker:
        Thread(target=runner.should_keep_running, daemon=True).start()
        log.info("started 'should_keep_running' thread")

    try:
        server.wait_for_termination()
    except Exception as e:
        log.error("server terminated with error: %s", e)
    finally:
        force_close(server)
