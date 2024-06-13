# List of functions/modules that should not run as activities.
# Function/modules added here should be:
# - Stateless
# - Common enough to be included
# So, `re` is OK, `random` is not

import datetime


def is_nonact(fn):
    """Return True if fn (callable) can run outside of activity."""
    return fn in functions or fn.__module__ in modules

# TODO:
# pathlib


# Modules are represented as strings func.__module__ is a string
modules = {
    'abc',
    'array',
    'base64',
    'bisect',
    'bz2',
    'cmath',
    'collections',
    'contextlib',
    'copy',
    'csv',
    'dataclasses',
    'decimal',
    'enum',
    'fnmatch',
    'fractions',
    'functools',
    'graphlib',
    'gzip',
    'hashlib',
    'heapq',
    'html',
    'html.entities',
    'html.parser',
    'ipaddress',
    'itertools',
    'json',
    'lzma',
    'math',
    'operator',
    'pprint',
    're',
    'shlex',
    'statistics',
    'stats',
    'struct',
    'tomllib',
    'traceback',
    'unicodedata',
    'zlib',
}

functions = {
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
}
