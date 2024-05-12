"""Run user code under AutoKitteh"""

# This file is long, but keeping it a single file helps with embedding it in Go and
# running it.

import ast
import builtins
import json
import logging
import pickle
import sys
import tarfile
from base64 import b64decode, b64encode
from functools import wraps
from importlib.abc import Loader
from importlib.machinery import SourceFileLoader
from os import mkdir
from pathlib import Path
from socket import AF_UNIX, SOCK_STREAM, socket

# Use own own logger, leave root logger to user.
log = logging.getLogger('AK')
log.setLevel(logging.INFO)
formatter = logging.Formatter('[%(name)s] %(asctime)s - %(levelname)s - %(filename)s:%(lineno)d - %(message)s')
hdlr = logging.StreamHandler()
hdlr.setFormatter(formatter)
log.addHandler(hdlr)

def name_of(node):
    """Name of call node (e.g. 'requests.get')"""
    if isinstance(node, ast.Attribute):
        prefix = name_of(node.value)
        return f'{prefix}.{node.attr}'

    if isinstance(node, ast.Call):
        return name_of(node.func)

    if isinstance(node, ast.Name):
        return node.id

    raise ValueError(f'unknown AST node type: {node!r}')


ACTION_NAME = '_ak_call'
MODULE_NAME = ''
BUILTIN = {v for v in dir(builtins) if callable(getattr(builtins, v))}


class Transformer(ast.NodeTransformer):
    """Replace 'fn(a, b)' with '_ak_call(fn, a, b)'."""
    def __init__(self, file_name):
        self.file_name = file_name

    def visit_Call(self, node):
        name = name_of(node.func)
        # ast.Transformer does not recurse to func or args
        node.func = self.visit(node.func)
        node.args = [self.visit(a) for a in node.args]

        if not name or name in BUILTIN:
            return node

        log.info('%s:%d: patching %s with action', self.file_name, node.lineno, name)
        call = ast.Call(
            func=ast.Name(id=ACTION_NAME, ctx=ast.Load()),
            args=[node.func] + node.args,
            keywords=node.keywords,
        )
        return call


class AKLoader(Loader):
    """Custom file loaders that will rewrite function calls to actions."""
    def __init__(self, src_loader, action):
        self.file_name = src_loader.path
        self.module_name = src_loader.name
        self.action = action

    def create_module(self, spec):
        # Must be defined since it's an abstract method
        return None  # Use default module

    def exec_module(self, module):
        try:
            with open(self.file_name) as fp:
                src = fp.read()
        except OSError as err:
            raise ImportError(f'cannot read {self.module_name!r} - {err}')

        mod = ast.parse(src, self.file_name, 'exec')
        trans = Transformer(self.file_name)
        out = trans.visit(mod)
        ast.fix_missing_locations(out)

        code = compile(out, self.file_name, 'exec')
        setattr(module, ACTION_NAME, self.action)
        exec(code, module.__dict__)


# There is an established way to add import hooks, but we want to change the behavior of
# the current PathFinder found in sys.import_hooks so it'll call our code when importing
# form the user directory. This is why you'll see all these monkey patches below.

def patch_finder(finder, action):
    """Patches the finder to use a custom source loader."""
    _orig_find_spec = finder.find_spec
    def find_spec(fullname, target=None):
        spec = _orig_find_spec(fullname, target)
        if spec is None or not isinstance(spec.loader, SourceFileLoader):
            return spec

        log.info('patching loader for %r', fullname)
        spec.loader = AKLoader(spec.loader, action)
        return spec

    finder.find_spec = find_spec


def wrap_hook(hook, user_dir, action):
    """Wraps a hook to patch finder on user code directories."""
    @wraps(hook)
    def wrapper(path):
        finder = hook(path)
        if user_dir.is_relative_to(path):
            patch_finder(finder, action)
        return finder

    return wrapper


def patch_import_hooks(user_dir, action_fn):
    """Patches standard import hook in sys.path_hooks."""
    user_dir = Path(user_dir)
    for i, hook in enumerate(sys.path_hooks):
        if hook.__name__ == 'path_hook_for_FileFinder':
            sys.path_hooks[i] = wrap_hook(hook, user_dir, action_fn)
            return

    raise RuntimeError(f'cannot find import hook to patch in {sys.path_hooks}')


def load_code(root_path, action_fn, module_name):
    patch_import_hooks(root_path, action_fn)
    sys.path.insert(0, str(root_path))
    log.info('importing %r', module_name)
    mod = __import__(module_name)
    return mod


def run_code(mod, entry_point, data):
    fn = getattr(mod, entry_point, None)
    if fn is None:
        raise NameError('%s.%s not found', mod.__name__, entry_point)

    log.info('calling %s.%s', mod.__name__, entry_point)
    return fn(data)


class MessageType:
    callback = 'callback'
    done = 'done'
    log = 'log'
    module = 'module'
    response = 'response'
    run = 'run'
    

class Comm:
    def __init__(self, sock):
        self.sock = sock
        self.rdr = sock.makefile('r')

    def _send(self, message):
        data = json.dumps(message) + '\n'
        self.sock.sendall(data.encode('utf-8'))

    def recv(self, *msg_types):
        data = self.rdr.readline()
        if not data:
            raise ValueError('connection closed')

        message = json.loads(data)
        if (typ := message['type']) not in msg_types:
            typs = ', '.join(msg_types)
            raise ValueError(f'message type: expected one of {typs!r}, got {typ!r}')
        return message

    def _picklize(self, data):
        data = pickle.dumps(data, protocol=0)
        return b64encode(data).decode('utf-8')

    def send_activity(self, fn, args, kw):
        data = (fn, args, kw)
        message = {
            'type': MessageType.callback,
            'payload': {
                'name': fn.__name__,
                'args': [str(a) for a in args],
                'kw': {k: str(v) for k, v in kw.items()},
                'data': self._picklize(data),
            },
        }
        self._send(message)

    def extract_activity(self, message):
        payload = message['payload']
        data = b64decode(payload['data'])
        payload['data'] = pickle.loads(data)
        return payload

    def send_exported(self, entries):
        message = {
            'type': MessageType.module,
            'payload': {
                'entries': entries,
            }
        }
        self._send(message)

    def send_done(self):
        message = {'type': MessageType.done}
        self._send(message)

    def receive_run(self):
        message = self.recv(MessageType.run)
        return message['payload']

    def send_response(self, value):
        message = {
            'type': MessageType.response,
            'payload': {
                'value': self._picklize(value),
            }
        }
        self._send(message)

    def extract_response(self, message):
        data = message['payload']['value']
        return pickle.loads(b64decode(data))


class AKCall:
    """Callable wrapping functions with activities."""
    def __init__(self, module_name, comm: Comm):
        self.module_name = module_name
        self.in_activity = False
        self.comm = comm

    def ignore(self, fn):
        if fn.__module__ == 'builtins':
            return True
        
        if fn.__module__ == self.module_name:
            return True

        return False

    def __call__(self, func, *args, **kw):
        if self.in_activity or self.ignore(func):
            log.info(
                'calling %s (args=%r, kw=%r) directly (in_activity=%s)', 
                func.__name__, args, kw, self.in_activity)
            return func(*args, **kw)

        log.info('ACTION: calling %s via activity (args=%r, kw=%r)', func.__name__, args, kw)
        self.in_activity = True
        try:
            self.comm.send_activity(func, args, kw)
            message = self.comm.recv(MessageType.callback, MessageType.response)
            
            if message['type'] == MessageType.callback:
                payload = self.comm.extract_activity(message)
                fn, args, kw = payload['data']
                value = fn(*args, **kw)
                self.comm.send_response(value)
                message = self.comm.recv(MessageType.response)

            # Reply message, either from current call or playback
            return self.comm.extract_response(message)
        finally:
            self.in_activity = False


def extract_code(tar_path):
    root_dir = Path(tar_path).absolute().parent
    code_dir = f'{root_dir}/code'
    mkdir(code_dir)
    with tarfile.open(tar_path) as tf:
        tf.extractall(code_dir)

    return code_dir


def parse_path(root_path):
    """
    >>> parse_path('review.py:on_github_pull_request')
    ('review', 'on_github_pull_request')
    """
    if ':' not in root_path:
        raise ValueError(f'{root_path!r} - missing :')

    file_name, func_name = root_path.split(':', 1)
    if not file_name.endswith('.py'):
        raise ValueError(f'{root_path!r} - not a Python file')

    return file_name[:-3], func_name


# argparse.FileType will open the file
def file_type(value):
    path = Path(value)
    if path.is_file() or path.is_socket():
        return value

    raise ValueError(f'{value!r} - not a file')


def encode_msg(typ, name, payload, func_name="", func_args=None, kw=None):
    if isinstance(payload, str):
        payload = payload.encode('utf-8')
    data = b64encode(payload)

    data = json.dumps({
        'type': typ,
        'name': name,
        'payload': data.decode('utf-8'),
        'func': {
            'name': func_name,
            'args': func_args or [],
            'kw': kw or {},
        },
    }) + '\n'
    return data.encode('utf-8')


def decode_msg(data):
    obj = json.loads(data)
    if obj.get('payload'):
        obj['payload'] = b64decode(obj['payload'])

    return obj


def module_entries(mod):
    return [
        name
        for name in dir(mod)
        if name != ACTION_NAME and callable(getattr(mod, name, None))
    ]


if __name__ == '__main__':
    from argparse import ArgumentParser
    import sys

    parser = ArgumentParser(description='autokitteh Python runner')
    parser.add_argument('sock', help='path to unix domain socket', type=file_type)
    parser.add_argument('tar', help='path to code tar file', type=file_type)
    parser.add_argument('path', help='file.py:function')
    args = parser.parse_args()

    # TODO: Ask Itay why AK does not pass entry point
    if ':' in args.path:
        module_name, _ = parse_path(args.path)
    else:
        module_name = args.path[:-3]

    py_version = '{}.{}'.format(*sys.version_info[:2])
    log.info('python: %r, version: %r', sys.executable, py_version)
    log.info('sock: %r, tar: %r, module: %r', args.sock, args.tar, module_name)
    code_dir = extract_code(args.tar)
    log.info('code dir: %r', code_dir)

    sock = socket(AF_UNIX, SOCK_STREAM)
    sock.connect(args.sock)
    log.info('connected to %r', args.sock)

    log.info('loading %r', module_name)
    comm = Comm(sock)

    ak_call = AKCall(module_name, comm)
    mod = load_code(code_dir, ak_call, module_name)
    MODULE_NAME = mod.__name__
    entries = module_entries(mod)
    comm.send_exported(entries)

    # Initial call.
    message = comm.receive_run()
    func_name = message.get('func_name')
    if func_name is None:
        log.error('no function name in %r', message)
        raise SystemExit(1)

    fn = getattr(mod, func_name, None)
    if fn is None:
        log.error('%r has no function %r', module_name, func_name)
        raise SystemExit(1)
        
    event = message.get('event')
    event = {} if event is None else event

    # Inject HTTP body
    # TODO (ENG-624) change this once we support callbacks to autokitteh
    body = event.get('data', {}).get('body')
    if isinstance(body, str):
        try:
            event['data']['body'] = b64decode(body)
        except ValueError:
            pass

    fn(event)
    comm.send_done()
