import inspect
from pathlib import Path
from time import sleep

import autokitteh
from autokitteh import decorators

from . import log
from .comm import Comm, MessageType
from .deterministic import is_deterministic

# Functions that are called back to ak
AK_FUNCS = {
    autokitteh.next_event,
    autokitteh.subscribe,
    autokitteh.unsubscribe,
    sleep,
}


def is_marked_activity(fn):
    """Return true if function is marked as an activity."""
    return getattr(fn, decorators.ACTIVITY_ATTR, False)


class AKCall:
    """Callable wrapping functions with activities."""

    def __init__(self, comm: Comm):
        self.comm = comm

        self.in_activity = False
        self.loading = True  # Loading module
        self.module = None  # Module for "local" function, filled by "run"

    def is_module_func(self, fn):
        return fn.__module__ == self.module.__name__

    def should_run_as_activity(self, fn):
        if self.in_activity or self.loading:
            return False

        if is_marked_activity(fn):
            return True

        if is_deterministic(fn):
            return False

        if self.is_module_func(fn):
            return False

        return True

    def set_module(self, mod):
        self.module = mod
        self.loading = False

    def __call__(self, func, *args, **kw):
        if not callable(func):
            frames = inspect.stack()
            if len(frames) > 1:
                file, lnum = Path(frames[1].filename).name, frames[1].lineno
            else:
                file, lnum = "<unknown>", 0

            raise ValueError(f"{func!r} is not callable (user bug at {file}:{lnum}?)")

        if func in AK_FUNCS:
            if self.in_activity and func is sleep:
                return func(*args, **kw)

            log.info("ak function call: %s(%r, %r)", func.__name__, args, kw)
            self.comm.send_call(func.__name__, args, kw)
            msg = self.comm.recv(MessageType.call_return)
            value = msg["payload"]["value"]
            if func is autokitteh.next_event:
                value = autokitteh.AttrDict(value)
            return value

        if not self.should_run_as_activity(func):
            log.info(
                "calling %s (args=%r, kw=%r) directly (in_activity=%s)",
                func.__name__,
                args,
                kw,
                self.in_activity,
            )
            return func(*args, **kw)

        log.info("ACTION: activity call %s(%r, %r)", func.__name__, args, kw)
        self.in_activity = True
        try:
            if self.is_module_func(func):
                # Pickle can't handle function from our loaded module
                func = func.__name__
            self.comm.send_activity(func, args, kw)
            message = self.comm.recv(MessageType.callback, MessageType.response)

            if message["type"] == MessageType.callback:
                payload = self.comm.extract_activity(message)
                fn, args, kw = payload["data"]
                if isinstance(fn, str):
                    fn = getattr(self.module, fn, None)
                    if fn is None:
                        mod_name = self.module.__name__
                        raise ValueError(f"function {fn!r} not found in {mod_name!r}")
                value = fn(*args, **kw)
                self.comm.send_response(value)
                message = self.comm.recv(MessageType.response)

            # Reply message, either from current call or playback
            return self.comm.extract_response(message)
        finally:
            self.in_activity = False
