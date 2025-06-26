import inspect
import sys
import time
from pathlib import Path

import autokitteh
from autokitteh import activities

import log
from deterministic import is_deterministic, is_no_activity


ak_mod_name = autokitteh.__name__


def activity_marker(fn):
    return getattr(fn, activities.ACTIVITY_ATTR, None)


def callable_name(fn):
    for attr in ("__qualname__", "__name__"):
        name = getattr(fn, attr, None)
        if name:
            return name

    if hasattr(fn, "__class__"):
        return fn.__class__.__name__

    return repr(fn)  # last resort


def full_func_name(fn):
    name = callable_name(fn)
    module = getattr(fn, "__module__", None)
    if module:
        return f"{module}.{name}"

    return name


def caller_info(code_dir):
    """Return first location in call stack that comes from code_dir."""
    for frame in inspect.stack():
        if Path(frame.filename).is_relative_to(code_dir):
            return Path(frame.filename).name, frame.lineno

    return "<unknown>", 0


def is_ak_func(fn):
    mod = getattr(fn, "__module__", None)
    if not mod:
        if (inst := getattr(fn, "__self__", None)) is not None:
            mod = inst.__class__.__module__

    if not mod:
        return False

    return mod == ak_mod_name or mod.startswith(ak_mod_name + ".")


class AKCall:
    """Callable wrapping functions with activities."""

    def __init__(self, runner, code_dir: Path):
        self.runner = runner
        self.code_dir = code_dir.resolve()

        self.in_activity = False
        self.activities_inhibitions = 0  # Can be stacked. 0 means no inhibition.
        self.loading = True  # Loading module
        self.module = None  # Module for "local" function, filled by "run"

    def is_module_func(self, fn):
        mod_name = getattr(fn, "__module__", None)
        if mod_name is None:
            # Example: TypeError('int required').__init__
            return False

        if mod_name == self.module.__name__:
            return True

        mod = sys.modules.get(mod_name)
        if not mod:
            return False

        file_name = getattr(mod, "__file__", None)
        if file_name is None:
            return False

        mod_dir = Path(file_name).resolve()
        return mod_dir.is_relative_to(self.code_dir)

    def should_run_as_activity(self, fn):
        if self.in_activity or self.loading or self.activities_inhibitions:
            return False

        mark = activity_marker(fn)
        if mark in (True, False):
            return mark

        if is_ak_func(fn):
            return False

        if is_deterministic(fn):
            return False

        if is_no_activity(fn):
            return False

        if self.is_module_func(fn):
            return False

        return True

    def set_module(self, mod):
        self.module = mod
        self.loading = False

    def __call__(self, func, *args, **kw):
        file, lnum = caller_info(self.code_dir)
        log.info("CALLER: %s:%d", file, lnum)
        if not callable(func):
            raise ValueError(f"{func!r} is not callable (user bug at {file}:{lnum}?)")

        if func is time.sleep:
            if (n := len(args)) != 1:
                raise TypeError(f"time.sleep takes exactly one argument ({n} given)")

            seconds = args[0]
            fn = time.sleep if self.in_activity else self.runner.syscalls.ak_sleep
            return fn(seconds)

        inhibit = getattr(func, activities.INHIBIT_ACTIVITIES_ATTR, False)
        if inhibit:
            self.activities_inhibitions += 1
            log.info(f"inhibiting activities: {self.activities_inhibitions}")

        full_name = full_func_name(func)
        if not self.should_run_as_activity(func):
            log.info(
                f"calling {full_name} directly (in_activity={self.in_activity}, inhibitions={self.activities_inhibitions})",
            )

            try:
                return func(*args, **kw)
            finally:
                if inhibit:
                    log.info(f"uninhibiting activities: {self.activities_inhibitions}")
                    self.activities_inhibitions -= 1

        log.info("ACTION: activity call %s", full_name)
        self.in_activity = True
        try:
            return self.runner.call_in_activity(func, args, kw)
        finally:
            self.in_activity = False

    async def async_call(self, func, *args, **kw):
        return await self(func, *args, **kw)
