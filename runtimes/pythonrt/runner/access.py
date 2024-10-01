"""Restrict access to secrets"""

import traceback
from functools import wraps
from pathlib import Path

import autokitteh
import autokitteh.github

ak_dir = Path(autokitteh.__file__).parent


def is_under_ak():
    for frame in traceback.extract_stack():
        file_dir = Path(frame.filename).parent
        if file_dir.is_relative_to(ak_dir):
            return True

    return False


class Proxy:
    """Wrap object so that access to attr will raise AttributeError.

    It will allow access if attribute is accessed from autokitteh SDK.
    """

    def __init__(self, obj, attr):
        self._obj = obj
        self._secret_attr = attr

    def __getattribute__(self, attr):
        if attr in ("_obj", "_secret_attr"):
            raise AttributeError(attr)

        secret_attr = object.__getattribute__(self, "_secret_attr")
        if attr == secret_attr and not is_under_ak():
            raise AttributeError(attr)

        obj = object.__getattribute__(self, "_obj")
        return getattr(obj, attr)


def hide(fn, path, attr):
    """Decorator that hides a secret attribute from return value of a function.

    path is a dot delimited string for nested values (e.g. '_http.credentials.secret').

    >>> from unittest.mock import Mock
    >>> def connect():
    ...     client = Mock()
    ...     client.auth.version = 2
    ...     client.auth.token = 's3cr3t'
    ...     return client
    ...
    >>> connect = hide(connect, 'auth', 'token')
    >>> c = connect()
    >>> print('version:', c.auth.version)
    version: 2
    >>> print('token:', c.auth.token)
    Traceback (most recent call last):
        ...
    AttributeError: token
    """

    path = path.split(".") if path else []

    @wraps(fn)
    def wrapper(*args, **kw):
        out = fn(*args, **kw)
        prev, obj = None, out
        for a in path:
            prev, obj = obj, getattr(obj, a)

        proxy = Proxy(obj, attr)
        if prev:
            setattr(prev, path[-1], proxy)
        else:
            out = proxy
        return out

    return wrapper


# TODO: gmail ...
_hides = [
    (autokitteh.github, "github_client", "_Github__requester.auth", "private_key"),
]


def hide_ak_secrets():
    for mod, fn_name, path, attr in _hides:
        fn = getattr(mod, fn_name)
        fn2 = hide(fn, path, attr)
        setattr(mod, fn_name, fn2)
