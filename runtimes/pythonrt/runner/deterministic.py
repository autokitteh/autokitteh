"""List of functions/modules that are deterministic and should not run as activities."""

# Function/modules added here should be:
# - Stateless
# - Common enough to be included
# So, `re` is OK, `random` is not

import asyncio
import datetime
import functools
import inspect
import json

from autokitteh import activities


def attr_or_none(obj, attr):
    return getattr(obj, attr, None)


def is_deterministic(fn):
    """Return True if fn (callable) can run outside of activity."""
    if fn in functions:
        return True

    if attr_or_none(fn, "__module__") in modules:
        return True

    if hasattr(fn, "__self__"):  # A bound method
        cls = fn.__self__.__class__
        if cls in builtin_types:
            return True

        mod = cls.__module__
        if mod != "builtins" and mod in modules:
            return True

    if cls := getattr(fn, "__objclass__", None):
        # We don't check isinstance on built-in types since the user can subclass
        # and then override a method. For example:
        # class List(list):
        #     def append(self, item):
        #         ...
        # And this append should run as an activity.
        # We don't add Exception to builtin_types since then we'll need to add all
        # built-in exceptions.
        if cls in builtin_types or issubclass(cls, Exception):
            return True

    return False


# https://stackoverflow.com/questions/3589311/get-defining-class-of-unbound-method-object-in-python-3/25959545#25959545
def fn_class(meth):
    if isinstance(meth, functools.partial):
        return fn_class(meth.func)

    if inspect.ismethod(meth) or (
        inspect.isbuiltin(meth)
        and attr_or_none(meth, "__self__")
        and attr_or_none(meth.__self__, "__class__")
    ):
        for cls in inspect.getmro(meth.__self__.__class__):
            if meth.__name__ in cls.__dict__:
                return cls
        meth = getattr(meth, "__func__", meth)  # __qualname__ parsing

    if inspect.isfunction(meth):
        # TextCalender.prmonth -> TextCalendar
        cls_name = meth.__qualname__.split(".")[0]
        cls = attr_or_none(inspect.getmodule(meth), cls_name)
        if isinstance(cls, type):
            return cls

    return attr_or_none(meth, "__objclass__")  # Descriptor objects


def is_no_activity(fn):
    no_act = activities._no_activity

    if fn in no_act:
        return True

    # Bound method
    if inspect.ismethod(fn):
        if (cls_fn := attr_or_none(fn, "__func__")) and cls_fn in no_act:
            return True

    # Descriptors
    if (cls := attr_or_none(fn, "__objclass__")) and (
        name := attr_or_none(fn, "__name__")
    ):
        if cls_fn := attr_or_none(cls, name):
            return cls_fn in no_act

    if cls := fn_class(fn):
        return cls in no_act

    return False


# Please keep the following sorted in alphabetical order.

builtin_types = {
    bytearray,
    bytes,
    dict,
    frozenset,
    list,
    memoryview,
    range,
    set,
    str,
    tuple,
}

# Modules are represented as strings func.__module__ is a string
modules = {
    "abc",
    "array",
    "base64",
    "bisect",
    "builtins",
    "bz2",
    "cmath",
    "collections",
    "contextlib",
    "copy",
    "csv",
    "dataclasses",
    "decimal",
    "enum",
    "fnmatch",
    "fractions",
    "functools",
    "graphlib",
    "gzip",
    "hashlib",
    "heapq",
    "html.entities",
    "html.parser",
    "html",
    "ipaddress",
    "itertools",
    "lzma",
    "math",
    "operator",
    "pprint",
    "re",
    "shlex",
    "statistics",
    "stats",
    "struct",
    "textwrap",
    "tomllib",
    "traceback",
    "types",
    "typing",
    "unicodedata",
    "urllib.error",
    "urllib.parse",
    "xml.dom.minidom",
    "xml.dom.pulldom",
    "xml.dom",
    "xml.etree.ElementTree",
    "xml.parsers.expat",
    "xml.sax",
    "xml",
    "zipfile",
    "zlib",
    "zoneninfo",
}

functions = builtin_types | {
    # asyncio
    asyncio.create_task,
    asyncio.current_task,
    asyncio.gather,
    asyncio.LifoQueue,
    asyncio.PriorityQueue,
    asyncio.Queue,
    asyncio.run,
    asyncio.run_coroutine_threadsafe,
    asyncio.Runner,
    asyncio.shield,
    asyncio.Task,
    asyncio.timeout,
    asyncio.wait,
    asyncio.wait_for,
    # datetime
    datetime.date,
    datetime.date.fromisocalendar,
    datetime.date.fromisoformat,
    datetime.date.fromordinal,
    datetime.date.fromtimestamp,
    datetime.date.isocalendar,
    datetime.date.isoformat,
    datetime.date.isoweekday,
    datetime.date.replace,
    datetime.date.strftime,
    datetime.date.timetuple,
    datetime.date.toordinal,
    datetime.date.weekday,
    datetime.datetime,
    datetime.datetime.astimezone,
    datetime.datetime.combine,
    datetime.datetime.ctime,
    datetime.datetime.date,
    datetime.datetime.dst,
    datetime.datetime.fromisocalendar,
    datetime.datetime.fromisoformat,
    datetime.datetime.fromordinal,
    datetime.datetime.fromtimestamp,
    datetime.datetime.isocalendar,
    datetime.datetime.isoformat,
    datetime.datetime.isoformat,
    datetime.datetime.isoweekday,
    datetime.datetime.isoweekday,
    datetime.datetime.replace,
    datetime.datetime.strftime,
    datetime.datetime.strftime,
    datetime.datetime.strptime,
    datetime.datetime.time,
    datetime.datetime.timestamp,
    datetime.datetime.timetuple,
    datetime.datetime.timetuple,
    datetime.datetime.toordinal,
    datetime.datetime.toordinal,
    datetime.datetime.tzname,
    datetime.datetime.utcfromtimestamp,
    datetime.datetime.utcoffset,
    datetime.datetime.utctimetuple,
    datetime.datetime.weekday,
    datetime.datetime.weekday,
    datetime.timedelta,
    datetime.timedelta.total_seconds,
    datetime.time.fromisoformat,
    datetime.time.isoformat,
    datetime.time.replace,
    datetime.time.utcoffset,
    datetime.time.dst,
    datetime.time.tzname,
    # json
    # json.dump & json.load work with files
    json.dumps,
    json.loads,
}
