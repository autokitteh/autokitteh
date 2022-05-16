from concurrent import futures
from logging import Logger, getLogger
from typing import Optional, Tuple

import grpc

from autokitteh.plugin import Plugin
from autokitteh.pluginsgrpcsvc import PluginsGRPCSvc


defaultLogger = getLogger()


def init_pluginssvc(
    srv: grpc.Server,
    plugins: list[Plugin],
    l: Logger = defaultLogger,
) -> PluginsGRPCSvc:
    l.info(f'serving plugins: {",".join(p.id for p in plugins)}')

    svc = PluginsGRPCSvc(plugins, l=l)
    svc.register_in_server(srv)
    return svc


def serve_plugins(
    plugins: list[Plugin],
    num_workers: int = 5,
    addr: str = '[::]:50051',
    creds: Optional[grpc.ServerCredentials] = None,
    l: Logger = defaultLogger,
) -> Tuple[grpc.Server, int]:
    srv = grpc.server(futures.ThreadPoolExecutor(max_workers=num_workers))

    init_pluginssvc(srv, plugins, l)

    if creds:
        port = srv.add_secure_port(addr, creds)
    else:
        port = srv.add_insecure_port(addr)

    srv.start()

    l.info(f'started {"secure" if creds else "insecure"} grpc server, addr={addr}')

    return srv, port
