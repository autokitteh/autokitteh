"""List of functions/modules that are deterministic and should not run as activities."""

# Function/modules added here should be:
# - Stateless
# - Common enough to be included
# So, `re` is OK, `random` is not

import datetime
import json


def is_deterministic(fn):
    """Return True if fn (callable) can run outside of activity."""
    if fn in functions:
        return True

    if getattr(fn, "__module__", None) in modules:
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
    # json.dump & json.load work with files
    json.dumps,
    json.loads,
}
