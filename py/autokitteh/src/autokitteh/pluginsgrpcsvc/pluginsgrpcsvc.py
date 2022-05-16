from typing import Any, Optional
import grpc
import logging
import time

import autokitteh.proto.pluginsprovidersvc.svc_pb2 as svc_pb
import autokitteh.proto.pluginsprovidersvc.svc_pb2_grpc as svc_pb_grpc

from autokitteh.api import PluginID, Value, Error
from autokitteh.plugin import Plugin, PluginInstance, CallException


def _log(f: Any) -> Any:
    def wrap(*args: list[Any], **kwargs: dict[str, Any]) -> Any:
        l = args[0]._l # type: ignore

        l.debug(f'{f.__name__}: request={args[1]}')

        t0 = time.time()

        try:
            resp = f(*args, **kwargs)
            t1 = time.time()
        except Exception as e:
            t1 = time.time()
            l.error(f'{f.__name__}: t={t1-t0} exception {e}')
            raise

        l.debug(f'{f.__name__}: t={t1-t0} response={args[1]}')

        return resp

    return wrap


class PluginsGRPCSvc(svc_pb_grpc.PluginsProviderServicer): # type: ignore
    _l: logging.Logger
    _plugins: dict[PluginID, PluginInstance]

    def __init__(self, plugins: list[Plugin], l: logging.Logger) -> None:
        self._l = l
        self._plugins = {p.id: PluginInstance(p) for p in plugins}

    def register_in_server(self, srv: grpc.Server) -> None:
        svc_pb_grpc.add_PluginsProviderServicer_to_server(self, srv)

    def _get(self, context: grpc.ServicerContext, id: str) -> PluginInstance:
        try:
            return self._plugins[PluginID(id)]
        except KeyError:
            context.abort(grpc.StatusCode.NOT_FOUND, 'plugin not found')

    @_log
    def List(
        self,
        request: svc_pb.ListRequest,
        context: grpc.ServicerContext,
    ) -> svc_pb.ListResponse:
        return svc_pb.ListResponse(ids=self._plugins.keys())

    @_log
    def Describe(
        self,
        request: svc_pb.DescribeRequest,
        context: grpc.ServicerContext,
    ) -> svc_pb.DescribeResponse:
        p = self._get(context, request.id)
        return svc_pb.DescribeResponse(desc=p.plugin.desc)

    @_log
    def GetValues(
        self,
        request: svc_pb.GetValuesRequest,
        context: grpc.ServicerContext,
    ) -> svc_pb.GetValuesResponse:
        p = self._get(context, request.id)

        names = request.names or p.plugin.members.keys()

        values: dict[str, Value] = {}

        for name in names:
            v = p.get_value(name)
            if v:
                values[name] = v.pb

        return svc_pb.GetValuesResponse(values=values)


    @_log
    def CallValue(
        self,
        request: svc_pb.CallValueRequest,
        context: grpc.ServicerContext,
    ) -> svc_pb.CallValueResponse:
        p = self._get(context, request.id)

        err: Optional[str] = None
        v: Optional[Value] = None

        try:
            v = p.call_value(
                Value(pb=request.value),
                [Value(pb=a) for a in request.args],
                {k: Value(pb=v) for k, v in request.kwargs.items()},
            )
        except CallException as e:
            err = str(e)

        return svc_pb.CallValueResponse(
            retval=v.pb if v else None,
            error=Error(msg=err, type='call') if err else None,
        )
