import click
import importlib
import logging
import os
import requests
import sys

from autokitteh.api import PluginID
from autokitteh.plugin import Plugin
from autokitteh.pluginssrv import serve_plugins

AK_PLUGIN_ID = os.environ.get('AK_PLUGIN_ID')
AK_PROC_READY_ADDR = os.environ.get('AK_PROC_READY_ADDRESS')
AK_READY_ID = os.environ.get('AK_PROC_READY_ID', f'{os.getpid()}')

@click.command()
@click.option('--debug/--no-debug', default=False)
@click.option('--debugger/--no-debugger', default=False)
@click.option('--addr', type=str, default='[::]:0' if AK_PROC_READY_ADDR else '[::]:50051')
@click.option('--num-workers', 'nworkers', type=int, default=5)
@click.option('--plugin-id', type=str, default=lambda: AK_PLUGIN_ID)
@click.option('--ready-addr', type=str, default=lambda: AK_PROC_READY_ADDR)
@click.option('--ready-id', type=str, default=lambda: AK_READY_ID)
@click.argument('plugin', nargs=-1)
def serve(
    debug: bool,
    debugger: bool,
    addr: str,
    nworkers: int,
    plugin_id: str,
    ready_addr: str,
    ready_id: str,
    plugin: list[str],
) -> None:
    if debugger:
        import pdb; pdb.set_trace()

    logging.basicConfig(level=logging.DEBUG if debug else logging.INFO)

    plugins: list[Plugin] = []

    for p in plugin:
        parts = p.split(':', 1)

        path, name = parts[0], parts[1] if len(parts) == 2 else 'Plugin'

        mod = importlib.import_module(path)

        try:
            plug = mod.__dict__[name]
        except KeyError:
            raise click.ClickException(f'module {path} does not have {name}')

        if type(plug) != Plugin:
            raise click.ClickException(f'{path}:{name} is not a plugin')

        if plugin_id:
            # Plugin does not have id specified or plugin's id is the same
            # as the plugin_id plugin name.
            if not plug.id or plug.id == plugin_id.split('.')[-1]:
                plug = plug._replace(id=PluginID(plugin_id))
            elif plug.id != plugin_id:
                continue

            if len(plugins) > 0:
                raise click.ClickException('AK_PLUGIN_ID env is set, but more than a single plugin matches')

        plugins.append(plug)

        logging.info(f'loaded {path}:{name} as {plug.id}')

    if not plugins:
        logging.warning('no plugins loaded')

    if plugin_id and len(plugins) != 1:
        raise click.ClickException('AK_PLUGIN_ID env is set, but no plugins matched')

    srv, port = serve_plugins(plugins, addr=addr, num_workers=nworkers)

    logging.info(f'listening on port {port}')

    if ready_addr:
        resp = requests.post(ready_addr, params={'id': ready_id}, data=f'localhost:{port}')
        if resp.status_code != 200:
            raise click.ClickException(f'http ready response not ok: {resp.status_code}')
        logging.info(f'anounced ready to {ready_addr} with id {ready_id}')

    srv.wait_for_termination()
