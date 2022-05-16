from typing import Optional, Any, Callable
import hashlib
import inspect
import uuid

import autokitteh.proto.values.values_pb2 as values

from autokitteh.plugin import Plugin
from autokitteh.api import Value, Func, AutoKittehException


class PluginInstanceException(AutoKittehException):
    pass


class WrongIssuerError(PluginInstanceException):
    def __init__(self) -> None:
        super().__init__('wrong issuer')


class UnknownCallError(PluginInstanceException):
    def __init__(self, v: Value) -> None:
        self.v = v
        super().__init__(f'unknown call {v}')


class CallException(PluginInstanceException):
    def __init__(self, exc: Exception) -> None:
        self.exc = exc
        super().__init__(f'call: {exc}')


class NotACallValueError(PluginInstanceException):
    def __init__(self) -> None:
        super().__init__('value is not a call value')


class CallParameterError(PluginInstanceException):
    def __init__(self, name: str, error: str) -> None:
        self.name = name
        self.error = error
        super().__init__(f'{name}: {error}')


class PluginInstance(object):
    class _Funcs(object):
        _funcs: dict[str, Callable[..., Any]]
        _issuer: str

        def __init__(self, issuer: str) -> None:
            self._funcs = {}
            self._issuer = issuer

        def _add(self, id: str, name: str, f: Callable[..., Any]) -> Value:
            self._funcs[id] = f
            return Value.init(call=values.Call(
                id=id,
                name=name,
                issuer=self._issuer,
                flags=None,
            ))

        def add_unique(self, name: str, f: Func) -> Value:
            return self._add(hashlib.sha1(name.encode()).hexdigest(), name, f)

        def add_dynamic(self, name: str, f: Func) -> Value:
            return self._add(uuid.uuid4().hex, name, f)

        def __getitem__(self, id: str) -> Callable[..., Any]:
            return self._funcs[id]

    _plugin: Plugin
    _funcs: _Funcs
    _members: dict[str, Any]

    @property
    def plugin(self) -> Plugin:
        return self._plugin

    def __init__(self, plugin: Plugin) -> None:
        self._plugin = plugin
        self._funcs = PluginInstance._Funcs(issuer=plugin.id)

        self._members = {}
        for k, v in plugin.members.items():
            self._members[k] = Value.wrap(v, self._funcs.add_unique)

    def get_value(self, name: str) -> Optional[Value]:
        return self._members.get(name)

    def call_value(self, v: Value, args: list[Value], kwargs: dict[str, Value]) -> Value:
        call = v.pb.call
        if not call:
            raise NotACallValueError()

        if call.issuer != self._plugin.id:
            raise WrongIssuerError()

        try:
            f = self._funcs[call.id]
        except KeyError:
            raise UnknownCallError(v)

        return self._call(f, call.name, args, kwargs)


    def _call(self, f: Callable[..., Any], name: str, args: list[Value], kwargs: dict[str, Value]) -> Value:
        fkwargs: dict[str, Any] = {k: v.unwrapped for k, v in kwargs.items()}
        fargs: list[Any] = [v.unwrapped for v in args]

        sig = inspect.signature(f)

        for k, v in sig.parameters.items():
            if k == '_func_to_value' and '_func_to_value' not in kwargs:
                fkwargs['_func_to_value'] = self._funcs.add_dynamic
            elif k == '_name' and '_name' not in kwargs:
                fkwargs['_name'] = name

        try:
            ret = f(*fargs, **fkwargs)
        except Exception as e:
            raise CallException(e)

        return Value.wrap(ret, func_to_value=self._funcs.add_dynamic)
