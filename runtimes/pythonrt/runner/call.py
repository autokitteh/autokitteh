import inspect
from pathlib import Path
from time import sleep

import autokitteh
from autokitteh import decorators
from autokitteh.connections import refresh_oauth

import log
from deterministic import is_deterministic

# Functions that are called back to ak
AK_FUNCS = {
    autokitteh.next_event,
    autokitteh.subscribe,
    autokitteh.unsubscribe,
    refresh_oauth,
    sleep,
}


def is_marked_activity(fn):
    """Return true if function is marked as an activity."""
    return getattr(fn, decorators.ACTIVITY_ATTR, False)


def full_func_name(fn):
    module = getattr(fn, "__module__", None)
    if module:
        return f"{module}.{fn.__name__}"

    return fn.__name__


class AKCall:
    """Callable wrapping functions with activities."""

    def __init__(self, runner):
        self.runner = runner

        self.in_activity = False
        self.loading = True  # Loading module
        self.module = None  # Module for "local" function, filled by "run"

    def is_module_func(self, fn):
        # TODO: Check for all funcs in user code directory
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

        log.info("__call__: %s, args=%r, kw=%r", full_func_name(func), args, kw)
        if func in AK_FUNCS:
            if self.in_activity and func is sleep:
                return func(*args, **kw)

            log.info("ak function call: %s(%r, %r)", func.__name__, args, kw)
            return self.runner.syscall(func, args, kw)

        full_name = full_func_name(func)
        if not self.should_run_as_activity(func):
            log.info(
                "calling %s directly (in_activity=%s)", full_name, self.in_activity
            )
            return func(*args, **kw)

        log.info("ACTION: activity call %s", full_name)
        self.in_activity = True
        try:
            return self.runner.call_in_activity(func, args, kw)
        finally:
            self.in_activity = False
